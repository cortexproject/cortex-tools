// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Code generated by "pdata/internal/cmd/pdatagen/main.go". DO NOT EDIT.
// To regenerate this file run "make genpdata".

package pmetric

import (
	otlpmetrics "go.opentelemetry.io/collector/pdata/internal/data/protogen/metrics/v1"
)

// Metric represents one metric as a collection of datapoints.
// See Metric definition in OTLP: https://github.com/open-telemetry/opentelemetry-proto/blob/main/opentelemetry/proto/metrics/v1/metrics.proto
//
// This is a reference type, if passed by value and callee modifies it the
// caller will see the modification.
//
// Must use NewMetric function to create new instances.
// Important: zero-initialized instance is not valid for use.
type Metric struct {
	orig *otlpmetrics.Metric
}

func newMetric(orig *otlpmetrics.Metric) Metric {
	return Metric{orig}
}

// NewMetric creates a new empty Metric.
//
// This must be used only in testing code. Users should use "AppendEmpty" when part of a Slice,
// OR directly access the member if this is embedded in another struct.
func NewMetric() Metric {
	return newMetric(&otlpmetrics.Metric{})
}

// MoveTo moves all properties from the current struct overriding the destination and
// resetting the current instance to its zero value
func (ms Metric) MoveTo(dest Metric) {
	*dest.orig = *ms.orig
	*ms.orig = otlpmetrics.Metric{}
}

// Name returns the name associated with this Metric.
func (ms Metric) Name() string {
	return ms.orig.Name
}

// SetName replaces the name associated with this Metric.
func (ms Metric) SetName(v string) {
	ms.orig.Name = v
}

// Description returns the description associated with this Metric.
func (ms Metric) Description() string {
	return ms.orig.Description
}

// SetDescription replaces the description associated with this Metric.
func (ms Metric) SetDescription(v string) {
	ms.orig.Description = v
}

// Unit returns the unit associated with this Metric.
func (ms Metric) Unit() string {
	return ms.orig.Unit
}

// SetUnit replaces the unit associated with this Metric.
func (ms Metric) SetUnit(v string) {
	ms.orig.Unit = v
}

// Type returns the type of the data for this Metric.
// Calling this function on zero-initialized Metric will cause a panic.
func (ms Metric) Type() MetricType {
	switch ms.orig.Data.(type) {
	case *otlpmetrics.Metric_Gauge:
		return MetricTypeGauge
	case *otlpmetrics.Metric_Sum:
		return MetricTypeSum
	case *otlpmetrics.Metric_Histogram:
		return MetricTypeHistogram
	case *otlpmetrics.Metric_ExponentialHistogram:
		return MetricTypeExponentialHistogram
	case *otlpmetrics.Metric_Summary:
		return MetricTypeSummary
	}
	return MetricTypeEmpty
}

// Gauge returns the gauge associated with this Metric.
//
// Calling this function when Type() != MetricTypeGauge returns an invalid
// zero-initialized instance of Gauge. Note that using such Gauge instance can cause panic.
//
// Calling this function on zero-initialized Metric will cause a panic.
func (ms Metric) Gauge() Gauge {
	v, ok := ms.orig.GetData().(*otlpmetrics.Metric_Gauge)
	if !ok {
		return Gauge{}
	}
	return newGauge(v.Gauge)
}

// SetEmptyGauge sets an empty gauge to this Metric.
//
// After this, Type() function will return MetricTypeGauge".
//
// Calling this function on zero-initialized Metric will cause a panic.
func (ms Metric) SetEmptyGauge() Gauge {
	val := &otlpmetrics.Gauge{}
	ms.orig.Data = &otlpmetrics.Metric_Gauge{Gauge: val}
	return newGauge(val)
}

// Sum returns the sum associated with this Metric.
//
// Calling this function when Type() != MetricTypeSum returns an invalid
// zero-initialized instance of Sum. Note that using such Sum instance can cause panic.
//
// Calling this function on zero-initialized Metric will cause a panic.
func (ms Metric) Sum() Sum {
	v, ok := ms.orig.GetData().(*otlpmetrics.Metric_Sum)
	if !ok {
		return Sum{}
	}
	return newSum(v.Sum)
}

// SetEmptySum sets an empty sum to this Metric.
//
// After this, Type() function will return MetricTypeSum".
//
// Calling this function on zero-initialized Metric will cause a panic.
func (ms Metric) SetEmptySum() Sum {
	val := &otlpmetrics.Sum{}
	ms.orig.Data = &otlpmetrics.Metric_Sum{Sum: val}
	return newSum(val)
}

// Histogram returns the histogram associated with this Metric.
//
// Calling this function when Type() != MetricTypeHistogram returns an invalid
// zero-initialized instance of Histogram. Note that using such Histogram instance can cause panic.
//
// Calling this function on zero-initialized Metric will cause a panic.
func (ms Metric) Histogram() Histogram {
	v, ok := ms.orig.GetData().(*otlpmetrics.Metric_Histogram)
	if !ok {
		return Histogram{}
	}
	return newHistogram(v.Histogram)
}

// SetEmptyHistogram sets an empty histogram to this Metric.
//
// After this, Type() function will return MetricTypeHistogram".
//
// Calling this function on zero-initialized Metric will cause a panic.
func (ms Metric) SetEmptyHistogram() Histogram {
	val := &otlpmetrics.Histogram{}
	ms.orig.Data = &otlpmetrics.Metric_Histogram{Histogram: val}
	return newHistogram(val)
}

// ExponentialHistogram returns the exponentialhistogram associated with this Metric.
//
// Calling this function when Type() != MetricTypeExponentialHistogram returns an invalid
// zero-initialized instance of ExponentialHistogram. Note that using such ExponentialHistogram instance can cause panic.
//
// Calling this function on zero-initialized Metric will cause a panic.
func (ms Metric) ExponentialHistogram() ExponentialHistogram {
	v, ok := ms.orig.GetData().(*otlpmetrics.Metric_ExponentialHistogram)
	if !ok {
		return ExponentialHistogram{}
	}
	return newExponentialHistogram(v.ExponentialHistogram)
}

// SetEmptyExponentialHistogram sets an empty exponentialhistogram to this Metric.
//
// After this, Type() function will return MetricTypeExponentialHistogram".
//
// Calling this function on zero-initialized Metric will cause a panic.
func (ms Metric) SetEmptyExponentialHistogram() ExponentialHistogram {
	val := &otlpmetrics.ExponentialHistogram{}
	ms.orig.Data = &otlpmetrics.Metric_ExponentialHistogram{ExponentialHistogram: val}
	return newExponentialHistogram(val)
}

// Summary returns the summary associated with this Metric.
//
// Calling this function when Type() != MetricTypeSummary returns an invalid
// zero-initialized instance of Summary. Note that using such Summary instance can cause panic.
//
// Calling this function on zero-initialized Metric will cause a panic.
func (ms Metric) Summary() Summary {
	v, ok := ms.orig.GetData().(*otlpmetrics.Metric_Summary)
	if !ok {
		return Summary{}
	}
	return newSummary(v.Summary)
}

// SetEmptySummary sets an empty summary to this Metric.
//
// After this, Type() function will return MetricTypeSummary".
//
// Calling this function on zero-initialized Metric will cause a panic.
func (ms Metric) SetEmptySummary() Summary {
	val := &otlpmetrics.Summary{}
	ms.orig.Data = &otlpmetrics.Metric_Summary{Summary: val}
	return newSummary(val)
}

// CopyTo copies all properties from the current struct overriding the destination.
func (ms Metric) CopyTo(dest Metric) {
	dest.SetName(ms.Name())
	dest.SetDescription(ms.Description())
	dest.SetUnit(ms.Unit())
	switch ms.Type() {
	case MetricTypeGauge:
		ms.Gauge().CopyTo(dest.SetEmptyGauge())
	case MetricTypeSum:
		ms.Sum().CopyTo(dest.SetEmptySum())
	case MetricTypeHistogram:
		ms.Histogram().CopyTo(dest.SetEmptyHistogram())
	case MetricTypeExponentialHistogram:
		ms.ExponentialHistogram().CopyTo(dest.SetEmptyExponentialHistogram())
	case MetricTypeSummary:
		ms.Summary().CopyTo(dest.SetEmptySummary())
	}

}
