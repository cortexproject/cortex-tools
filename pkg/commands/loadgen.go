package commands

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
	"github.com/prometheus/prometheus/storage/remote"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	// 15 seconds is a common send interval. To provide useful metrics at high latencies we will
	// add 15 to the default Prometheus buckets
	defBuckets = append(prometheus.DefBuckets, 15)
)

var writeRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "loadgen",
	Name:      "write_request_duration_seconds",
	Buckets:   defBuckets,
}, []string{"success"})

var queryRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "loadgen",
	Name:      "query_request_duration_seconds",
	Buckets:   defBuckets,
}, []string{"success"})

type LoadgenCommand struct {
	writeURL       string
	activeSeries   int
	scrapeInterval time.Duration
	parallelism    int
	batchSize      int
	writeTimeout   time.Duration
	metricName     string

	queryURL         string
	query            string
	queryParallelism int
	queryTimeout     time.Duration
	queryDuration    time.Duration

	metricsListenAddress string

	// Runtime stuff.
	wg          sync.WaitGroup
	writeClient remote.WriteClient
	queryClient v1.API
}

func (c *LoadgenCommand) Register(app *kingpin.Application) {
	loadgenCommand := &LoadgenCommand{}
	cmd := app.Command("loadgen", "Simple load generator for Cortex.").Action(loadgenCommand.run)
	cmd.Flag("write-url", "").
		Default("").StringVar(&loadgenCommand.writeURL)
	cmd.Flag("series-name", "name of the metric that will be generated").
		Default("node_cpu_seconds_total").StringVar(&loadgenCommand.metricName)
	cmd.Flag("active-series", "number of active series to send").
		Default("1000").IntVar(&loadgenCommand.activeSeries)
	cmd.Flag("scrape-interval", "period to send metrics").
		Default("15s").DurationVar(&loadgenCommand.scrapeInterval)
	cmd.Flag("parallelism", "how many metrics to send simultaneously").
		Default("10").IntVar(&loadgenCommand.parallelism)
	cmd.Flag("batch-size", "how big a batch to send").
		Default("100").IntVar(&loadgenCommand.batchSize)
	cmd.Flag("write-timeout", "timeout for write requests").
		Default("500ms").DurationVar(&loadgenCommand.writeTimeout)

	cmd.Flag("query-url", "").
		Default("").StringVar(&loadgenCommand.queryURL)
	cmd.Flag("query", "query to run").
		Default("sum(node_cpu_seconds_total)").StringVar(&loadgenCommand.query)
	cmd.Flag("query-parallelism", "number of queries to run in parallel").
		Default("10").IntVar(&loadgenCommand.queryParallelism)
	cmd.Flag("query-timeout", "").
		Default("20s").DurationVar(&loadgenCommand.queryTimeout)
	cmd.Flag("query-duration", "length of query").
		Default("1h").DurationVar(&loadgenCommand.queryDuration)

	cmd.Flag("metrics-listen-address", "address to serve metrics on").
		Default(":8080").StringVar(&loadgenCommand.metricsListenAddress)
}

func (c *LoadgenCommand) run(_ *kingpin.ParseContext) error {
	if c.writeURL == "" && c.queryURL == "" {
		return errors.New("either a -write-url or -query-url flag must be provided to run the loadgen command")
	}

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		err := http.ListenAndServe(c.metricsListenAddress, nil)
		if err != nil {
			logrus.WithError(err).Errorln("metrics listener failed")
		}
	}()

	if c.writeURL != "" {
		log.Printf("setting up write load gen:\n  url=%s\n  parallelism: %v\n  active_series: %d\n interval: %v\n", c.writeURL, c.parallelism, c.activeSeries, c.scrapeInterval)
		writeURL, err := url.Parse(c.writeURL)
		if err != nil {
			return err
		}

		writeClient, err := remote.NewWriteClient("loadgen", &remote.ClientConfig{
			URL:     &config.URL{URL: writeURL},
			Timeout: model.Duration(c.writeTimeout),
		})
		if err != nil {
			return err
		}
		c.writeClient = writeClient

		c.wg.Add(c.parallelism)

		metricsPerShard := c.activeSeries / c.parallelism
		for i := 0; i < c.activeSeries; i += metricsPerShard {
			go c.runWriteShard(i, i+metricsPerShard)
		}
	} else {
		log.Println("write load generation is disabled, -write-url flag has not been set")
	}

	if c.queryURL != "" {
		log.Printf("setting up query load gen:\n  url=%s\n  parallelism: %v\n  query: %s", c.queryURL, c.queryParallelism, c.query)
		queryClient, err := api.NewClient(api.Config{
			Address: c.queryURL,
		})
		if err != nil {
			return err
		}
		c.queryClient = v1.NewAPI(queryClient)

		c.wg.Add(c.queryParallelism)

		for i := 0; i < c.queryParallelism; i++ {
			go c.runQueryShard()
		}
	} else {
		log.Println("query load generation is disabled, -query-url flag has not been set")
	}

	c.wg.Wait()
	return nil
}

func (c *LoadgenCommand) runWriteShard(from, to int) {
	defer c.wg.Done()
	ticker := time.NewTicker(c.scrapeInterval)
	c.runScrape(from, to)
	for range ticker.C {
		c.runScrape(from, to)
	}
}

func (c *LoadgenCommand) runScrape(from, to int) {
	for i := from; i < to; i += c.batchSize {
		if err := c.runBatch(i, i+c.batchSize); err != nil {
			log.Printf("error sending batch: %v", err)
		}
	}
	fmt.Printf("sent %d samples\n", to-from)
}

func (c *LoadgenCommand) runBatch(from, to int) error {
	var (
		req = prompb.WriteRequest{
			Timeseries: make([]prompb.TimeSeries, 0, to-from),
		}
		now = time.Now().UnixNano() / int64(time.Millisecond)
	)

	for i := from; i < to; i++ {
		timeseries := prompb.TimeSeries{
			Labels: []prompb.Label{
				{Name: "__name__", Value: c.metricName},
				{Name: "job", Value: "node_exporter"},
				{Name: "instance", Value: fmt.Sprintf("instance%000d", i)},
				{Name: "cpu", Value: "0"},
				{Name: "mode", Value: "idle"},
			},
			Samples: []prompb.Sample{{
				Timestamp: now,
				Value:     rand.Float64(),
			}},
		}
		req.Timeseries = append(req.Timeseries, timeseries)
	}

	data, err := proto.Marshal(&req)
	if err != nil {
		return err
	}

	compressed := snappy.Encode(nil, data)

	start := time.Now()
	attempt := 0 // TODO: do retries
	if _, err := c.writeClient.Store(context.Background(), compressed, attempt); err != nil {
		writeRequestDuration.WithLabelValues("error").Observe(time.Since(start).Seconds())
		return err
	}
	writeRequestDuration.WithLabelValues("success").Observe(time.Since(start).Seconds())

	return nil
}

func (c *LoadgenCommand) runQueryShard() {
	defer c.wg.Done()
	for {
		c.runQuery()
	}
}

func (c *LoadgenCommand) runQuery() {
	ctx, cancel := context.WithTimeout(context.Background(), c.queryTimeout)
	defer cancel()
	r := v1.Range{
		Start: time.Now().Add(-c.queryDuration),
		End:   time.Now(),
		Step:  time.Minute,
	}
	start := time.Now()
	_, _, err := c.queryClient.QueryRange(ctx, c.query, r)
	if err != nil {
		queryRequestDuration.WithLabelValues("error").Observe(time.Since(start).Seconds())
		log.Printf("error doing query: %v", err)
		return
	}
	queryRequestDuration.WithLabelValues("success").Observe(time.Since(start).Seconds())
}
