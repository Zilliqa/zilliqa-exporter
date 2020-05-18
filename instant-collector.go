package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

//Define a struct for you collector that contains pointers
//to prometheus descriptors for each metric you wish to expose.
//Note you can also include fields of other types if they provide utility
//but we just won't be exposing them as metrics.
type fooCollector struct {
	fooMetric *prometheus.Desc
	barMetric *prometheus.Desc
}

//You must create a constructor for you collector that
//initializes every descriptor and returns a pointer to the collector
func newInstantCollector() *fooCollector {
	return &fooCollector{
		fooMetric: prometheus.NewDesc("unix",
			"Shows whether a foo has occurred in our cluster",
			nil, nil,
		),
		barMetric: prometheus.NewDesc("unix_nano",
			"Shows whether a bar has occurred in our cluster",
			nil, nil,
		),
	}
}

//Each and every collector must implement the Describe function.
//It essentially writes all descriptors to the prometheus desc channel.
func (collector *fooCollector) Describe(ch chan<- *prometheus.Desc) {

	//Update this section with the each metric you create for a given collector
	ch <- collector.fooMetric
	ch <- collector.barMetric
}

//Collect implements required collect function for all promehteus collectors
func (collector *fooCollector) Collect(ch chan<- prometheus.Metric) {

	//Implement logic here to determine proper metric value to return to prometheus
	//for each descriptor or call other functions that do so.
	now := time.Now()

	//Write latest value for each metric in the prometheus metric channel.
	//Note that you can pass CounterValue, GaugeValue, or UntypedValue types here.
	ch <- prometheus.MustNewConstMetric(collector.fooMetric, prometheus.GaugeValue, float64(now.Unix()))
	ch <- prometheus.MustNewConstMetric(collector.barMetric, prometheus.GaugeValue, float64(now.UnixNano()))

}
