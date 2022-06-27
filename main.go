package main

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/promql"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/tsdb/chunkenc"
)

type queryable struct{}

func newQueryable() storage.Queryable {
	return &queryable{}
}

func (queryable) Querier(ctx context.Context, _, _ int64) (storage.Querier, error) {
	return newQuerier(ctx), nil
}

type querier struct {
	ctx context.Context
}

func newQuerier(ctx context.Context) storage.Querier {
	return &querier{ctx: ctx}
}

func (querier) Select(
	sortSeries bool,
	hints *storage.SelectHints,
	labelMatchers ...*labels.Matcher,
) storage.SeriesSet {
	labels := labels.Labels{
		labels.Label{
			Name:  "__name__",
			Value: "bar",
		},
	}
	samples := []sample{}

	currValue := float64(0)
	for ts := hints.Start; ts <= hints.End; ts += 30000 {
		samples = append(samples, sample{
			timestamp: ts,
			value:     currValue,
		})
		currValue += 5
	}

	fmt.Println("---Input:")
	for _, sample := range samples {
		fmt.Printf("%v => %v\n", time.Unix(sample.timestamp/1000, 0), sample.value)
	}

	return newSeriesSet([]storage.Series{
		newSeries(labels, samples),
	})
}

func (querier) Close() error {
	return nil
}

func (querier) LabelNames() ([]string, storage.Warnings, error) {
	panic("1 not implemented")
}

func (querier) LabelValues(name string, matchers ...*labels.Matcher) ([]string, storage.Warnings, error) {
	panic("2 not implemented")
}

type sample struct {
	// Epoch timestamp in millisecond because promql works in milliseconds.
	timestamp int64
	value     float64
}

type series struct {
	labels  labels.Labels
	samples []sample
}

func newSeries(labels labels.Labels, samples []sample) storage.Series {
	return &series{
		labels:  labels,
		samples: samples,
	}
}

func (s *series) Labels() labels.Labels {
	return s.labels
}

func (s *series) Iterator() chunkenc.Iterator {
	return newSampleIterator(s.samples)
}

type sampleIterator struct {
	ix      int
	samples []sample
}

func newSampleIterator(samples []sample) chunkenc.Iterator {
	return &sampleIterator{
		ix:      -1,
		samples: samples,
	}
}

func (it *sampleIterator) Next() bool {
	it.ix++
	if it.ix < len(it.samples) {
		return true
	}

	return false
}

func (it *sampleIterator) Seek(_ int64) bool {
	return true
}

func (it *sampleIterator) At() (int64, float64) {
	return it.samples[it.ix].timestamp, it.samples[it.ix].value
}

func (it *sampleIterator) Err() error {
	return nil
}

type seriesSet struct {
	ix     int
	series []storage.Series
}

func newSeriesSet(series []storage.Series) storage.SeriesSet {
	return &seriesSet{
		ix:     -1,
		series: series,
	}
}

func (s *seriesSet) Next() bool {
	s.ix++
	if s.ix < len(s.series) {
		return true
	}

	return false
}

func (s *seriesSet) At() storage.Series {
	return s.series[s.ix]
}

func (s *seriesSet) Err() error {
	return nil
}

func (s *seriesSet) Warnings() storage.Warnings {
	return nil
}

func main() {
	engine := promql.NewEngine(promql.EngineOpts{
		MaxSamples:           1000000,
		Timeout:              5 * time.Minute,
		LookbackDelta:        5 * time.Minute, // Default value
		EnableAtModifier:     true,
		EnableNegativeOffset: true,
	})

	end := time.Now().Truncate(30 * time.Second)
	start := end.Add(-10 * time.Minute)
	step := 15 * time.Second

	query, err := engine.NewRangeQuery(
		newQueryable(),
		"rate(bar[1m])",
		start,
		end,
		step,
	)
	if err != nil {
		panic(err)
	}

	res := query.Exec(context.Background())
	matrix, err := res.Matrix()
	if err != nil {
		panic(err)
	}

	series := matrix[0]
	fmt.Println("\n---Output")
	fmt.Println(series.Metric.String())
	for _, sample := range series.Points {
		fmt.Printf("%v => %v\n", time.Unix(sample.T/1000, 0), sample.V)
	}
}
