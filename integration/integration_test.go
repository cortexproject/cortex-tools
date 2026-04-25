//go:build integration

package integration

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/prometheus/prometheus/model/rulefmt"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

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
