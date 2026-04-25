//go:build integration

package integration

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	config_util "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/rulefmt"
	"github.com/prometheus/prometheus/prompb"
	"github.com/prometheus/prometheus/promql/parser"
	"github.com/prometheus/prometheus/storage/remote"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/cortexproject/cortex-tools/pkg/backfill"
	"github.com/cortexproject/cortex-tools/pkg/bench"
	"github.com/cortexproject/cortex-tools/pkg/client"
	"github.com/cortexproject/cortex-tools/pkg/rules/rwrulefmt"
)

func cortexAddress() string {
	if addr := os.Getenv("CORTEX_ADDRESS"); addr != "" {
		return addr
	}
	return "http://localhost:9009"
}

func newClient(t *testing.T) *client.CortexClient {
	t.Helper()
	c, err := client.New(client.Config{
		Address: cortexAddress(),
		ID:      "fake",
	})
	require.NoError(t, err)
	return c
}

func ruleNode(record, expr string) rulefmt.RuleNode {
	return rulefmt.RuleNode{
		Record: yaml.Node{Kind: yaml.ScalarNode, Value: record},
		Expr:   yaml.Node{Kind: yaml.ScalarNode, Value: expr},
	}
}

func TestRulesLoadListDelete(t *testing.T) {
	ctx := context.Background()
	c := newClient(t)

	namespace := "test_namespace"
	group := rwrulefmt.RuleGroup{}
	group.Name = "test_rule_group"
	group.Rules = []rulefmt.RuleNode{
		ruleNode("summed_up", "sum(up)"),
	}

	err := c.CreateRuleGroup(ctx, namespace, group)
	require.NoError(t, err, "CreateRuleGroup should succeed")

	ruleSet, err := c.ListRules(ctx, "")
	require.NoError(t, err, "ListRules should succeed")
	require.Contains(t, ruleSet, namespace, "namespace should exist")

	found := false
	for _, g := range ruleSet[namespace] {
		if g.Name == "test_rule_group" {
			found = true
			break
		}
	}
	require.True(t, found, "rule group should be in list")

	rg, err := c.GetRuleGroup(ctx, namespace, "test_rule_group")
	require.NoError(t, err, "GetRuleGroup should succeed")
	require.Equal(t, "test_rule_group", rg.Name)
	require.Len(t, rg.Rules, 1)

	err = c.DeleteRuleNamespace(ctx, namespace)
	require.NoError(t, err, "DeleteRuleNamespace should succeed")

	ruleSet, err = c.ListRules(ctx, "")
	if err != nil {
		require.True(t, errors.Is(err, client.ErrResourceNotFound), "expected no rules or resource not found, got: %v", err)
	} else {
		require.NotContains(t, ruleSet, namespace, "namespace should be deleted")
	}
}

func TestRulesMultipleGroups(t *testing.T) {
	ctx := context.Background()
	c := newClient(t)

	namespace := "multi_group_namespace"
	groups := []rwrulefmt.RuleGroup{
		{RuleGroup: rulefmt.RuleGroup{Name: "group_a", Rules: []rulefmt.RuleNode{ruleNode("metric_a", "sum(up)")}}},
		{RuleGroup: rulefmt.RuleGroup{Name: "group_b", Rules: []rulefmt.RuleNode{ruleNode("metric_b", "count(up)")}}},
		{RuleGroup: rulefmt.RuleGroup{Name: "group_c", Rules: []rulefmt.RuleNode{ruleNode("metric_c", "avg(up)")}}},
	}

	for _, g := range groups {
		err := c.CreateRuleGroup(ctx, namespace, g)
		require.NoError(t, err, "CreateRuleGroup %s should succeed", g.Name)
	}

	ruleSet, err := c.ListRules(ctx, namespace)
	require.NoError(t, err)
	require.Contains(t, ruleSet, namespace)
	require.Len(t, ruleSet[namespace], 3, "should have 3 rule groups")

	err = c.DeleteRuleNamespace(ctx, namespace)
	require.NoError(t, err)
}

func TestAlertmanagerLoadGet(t *testing.T) {
	ctx := context.Background()
	c := newClient(t)

	amConfig := "route:\n  receiver: default\nreceivers:\n  - name: default\n"

	err := c.CreateAlertmanagerConfig(ctx, amConfig, nil)
	require.NoError(t, err, "CreateAlertmanagerConfig should succeed")

	cfg, templates, err := c.GetAlertmanagerConfig(ctx)
	require.NoError(t, err, "GetAlertmanagerConfig should succeed")
	require.Contains(t, cfg, "receiver: default")
	require.Empty(t, templates)

	err = c.DeleteAlermanagerConfig(ctx)
	require.NoError(t, err, "DeleteAlermanagerConfig should succeed")
}

func TestAlertmanagerWithTemplates(t *testing.T) {
	ctx := context.Background()
	c := newClient(t)

	amConfig := "route:\n  receiver: default\nreceivers:\n  - name: default\n"
	tmpl := map[string]string{
		"slack.tmpl": "{{ define \"slack.title\" }}Alert: {{ .CommonLabels.alertname }}{{ end }}",
	}

	err := c.CreateAlertmanagerConfig(ctx, amConfig, tmpl)
	require.NoError(t, err, "CreateAlertmanagerConfig with templates should succeed")

	cfg, templates, err := c.GetAlertmanagerConfig(ctx)
	require.NoError(t, err, "GetAlertmanagerConfig should succeed")
	require.Contains(t, cfg, "receiver: default")
	require.Contains(t, templates, "slack.tmpl")
	require.Contains(t, templates["slack.tmpl"], "slack.title")

	err = c.DeleteAlermanagerConfig(ctx)
	require.NoError(t, err, "DeleteAlermanagerConfig should succeed")

	// Verify config is gone
	_, _, err = c.GetAlertmanagerConfig(ctx)
	require.Error(t, err, "GetAlertmanagerConfig should fail after delete")
}

func remoteWrite(t *testing.T, series []prompb.TimeSeries) {
	t.Helper()

	writeReq := &prompb.WriteRequest{Timeseries: series}
	data, err := proto.Marshal(writeReq)
	require.NoError(t, err)

	compressed := snappy.Encode(nil, data)

	writeURL := cortexAddress() + "/api/v1/push"
	req, err := http.NewRequest("POST", writeURL, bytes.NewReader(compressed))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/x-protobuf")
	req.Header.Set("Content-Encoding", "snappy")
	req.Header.Set("X-Prometheus-Remote-Write-Version", "0.1.0")
	req.Header.Set("X-Scope-OrgID", "fake")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, "remote write should return 200")
}

func TestRemoteRead(t *testing.T) {
	now := time.Now()

	series := []prompb.TimeSeries{
		{
			Labels: []prompb.Label{
				{Name: "__name__", Value: "integration_test_metric"},
				{Name: "job", Value: "test"},
			},
			Samples: []prompb.Sample{
				{Value: 42.0, Timestamp: now.UnixMilli()},
				{Value: 43.0, Timestamp: now.Add(-30 * time.Second).UnixMilli()},
				{Value: 44.0, Timestamp: now.Add(-60 * time.Second).UnixMilli()},
			},
		},
	}

	remoteWrite(t, series)

	_, result := remoteReadQuery(t, `integration_test_metric`, now.Add(-5*time.Minute), now.Add(time.Minute))
	require.NotEmpty(t, result.Timeseries, "should have at least one timeseries")

	ts := result.Timeseries[0]
	foundMetricName := false
	for _, l := range ts.Labels {
		if l.Name == "__name__" && l.Value == "integration_test_metric" {
			foundMetricName = true
		}
	}
	require.True(t, foundMetricName, "timeseries should have the correct metric name")
	require.GreaterOrEqual(t, len(ts.Samples), 1, "should have samples")

	fmt.Printf("Remote read returned %d timeseries with %d samples\n", len(result.Timeseries), len(ts.Samples))
}

func TestRemoteReadExport(t *testing.T) {
	now := time.Now()

	series := []prompb.TimeSeries{
		{
			Labels: []prompb.Label{
				{Name: "__name__", Value: "export_test_metric"},
				{Name: "job", Value: "export_test"},
			},
			Samples: []prompb.Sample{
				{Value: 100.0, Timestamp: now.UnixMilli()},
			},
		},
	}

	remoteWrite(t, series)

	// Read the data back
	readClient, result := remoteReadQuery(t, `export_test_metric`, now.Add(-5*time.Minute), now.Add(time.Minute))
	_ = readClient
	require.NotEmpty(t, result.Timeseries)

	// Export to TSDB using backfill.CreateBlocks
	tsdbPath, err := os.MkdirTemp("", "cortex-integration-export-*")
	require.NoError(t, err)
	defer os.RemoveAll(tsdbPath)

	from := now.Add(-5 * time.Minute)
	to := now.Add(time.Minute)
	mint := int64(model.TimeFromUnixNano(from.UnixNano()))
	maxt := int64(model.TimeFromUnixNano(to.UnixNano()))

	timeseries := result.Timeseries
	iterator := func() backfill.Iterator {
		return newTimeSeriesIterator(timeseries)
	}

	err = backfill.CreateBlocks(iterator, mint, maxt, 1000, tsdbPath, false, io.Discard)
	require.NoError(t, err, "CreateBlocks should succeed")

	// Verify TSDB blocks were created
	entries, err := os.ReadDir(tsdbPath)
	require.NoError(t, err)
	blockCount := 0
	for _, e := range entries {
		if e.IsDir() && e.Name() != "wal" {
			blockCount++
		}
	}
	require.GreaterOrEqual(t, blockCount, 1, "should have created at least one TSDB block")
	fmt.Printf("Export created %d TSDB blocks in %s\n", blockCount, tsdbPath)
}

func TestLoadgen(t *testing.T) {
	now := time.Now()

	seriesDescs := []bench.SeriesDesc{
		{
			Name: "loadgen_test_metric",
			Type: bench.GaugeRandom,
			Labels: []bench.LabelDesc{
				{Name: "instance", ValuePrefix: "inst", UniqueValues: 2},
			},
		},
	}

	series, totalSeriesTypeMap := bench.SeriesDescToSeries(seriesDescs)
	totalSeries := 0
	for _, n := range totalSeriesTypeMap {
		totalSeries += n
	}

	workload := &bench.WriteWorkload{
		Replicas:           1,
		Series:             series,
		TotalSeries:        totalSeries,
		TotalSeriesTypeMap: totalSeriesTypeMap,
	}

	timeseries := workload.GenerateTimeSeries("integration-test", now)
	require.NotEmpty(t, timeseries, "workload should generate timeseries")

	remoteWrite(t, timeseries)

	_, result := remoteReadQuery(t, `{__name__="loadgen_test_metric"}`, now.Add(-5*time.Minute), now.Add(time.Minute))
	require.NotEmpty(t, result.Timeseries, "should read back loadgen timeseries")

	fmt.Printf("Loadgen wrote %d series, read back %d timeseries\n", len(timeseries), len(result.Timeseries))
}

func remoteReadQuery(t *testing.T, selector string, from, to time.Time) (remote.ReadClient, *prompb.QueryResult) {
	t.Helper()

	addressURL, err := url.Parse(cortexAddress())
	require.NoError(t, err)
	addressURL.Path = filepath.Join(addressURL.Path, "/prometheus/api/v1/read")

	readClient, err := remote.NewReadClient("test", &remote.ClientConfig{
		URL:     &config_util.URL{URL: addressURL},
		Timeout: model.Duration(30 * time.Second),
	})
	require.NoError(t, err)

	rc := readClient.(*remote.Client)
	rc.Client.Transport = &tenantIDTransport{
		RoundTripper: http.DefaultTransport,
		tenantID:     "fake",
	}

	matchers, err := parser.ParseMetricSelector(selector)
	require.NoError(t, err)

	pbQuery, err := remote.ToQuery(
		int64(model.TimeFromUnixNano(from.UnixNano())),
		int64(model.TimeFromUnixNano(to.UnixNano())),
		matchers,
		nil,
	)
	require.NoError(t, err)

	result, err := readClient.Read(context.Background(), pbQuery)
	require.NoError(t, err)
	return readClient, result
}

func newTimeSeriesIterator(ts []*prompb.TimeSeries) *timeSeriesIterator {
	return &timeSeriesIterator{
		posSeries:       0,
		posSample:       -1,
		labelsSeriesPos: -1,
		ts:              ts,
	}
}

type timeSeriesIterator struct {
	posSeries       int
	posSample       int
	ts              []*prompb.TimeSeries
	labels          labels.Labels
	labelsSeriesPos int
}

func (i *timeSeriesIterator) Next() error {
	if i.posSeries >= len(i.ts) {
		return io.EOF
	}
	i.posSample++
	if i.posSample >= len(i.ts[i.posSeries].Samples) {
		i.posSample = -1
		i.posSeries++
		return i.Next()
	}
	return nil
}

func (i *timeSeriesIterator) Labels() labels.Labels {
	if i.posSeries == i.labelsSeriesPos {
		return i.labels
	}
	series := i.ts[i.posSeries]
	i.labels = make(labels.Labels, len(series.Labels))
	for idx := range series.Labels {
		i.labels[idx].Name = series.Labels[idx].Name
		i.labels[idx].Value = series.Labels[idx].Value
	}
	i.labelsSeriesPos = i.posSeries
	return i.labels
}

func (i *timeSeriesIterator) Sample() (int64, float64) {
	series := i.ts[i.posSeries]
	sample := series.Samples[i.posSample]
	return sample.GetTimestamp(), sample.GetValue()
}

type tenantIDTransport struct {
	http.RoundTripper
	tenantID string
}

func (t *tenantIDTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("X-Scope-OrgID", t.tenantID)
	return t.RoundTripper.RoundTrip(req)
}
