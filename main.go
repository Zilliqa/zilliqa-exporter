package main

import (
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetOutput(os.Stderr)
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	log.SetLevel(log.InfoLevel)
	if os.Getenv("DEBUG") != "" {
		log.SetLevel(log.DebugLevel)
	}

	//Create a new instance of the foo collector and
	//register it with the prometheus client.
	prometheus.MustRegister(newInstantCollector())

	gauge := prometheus.NewGauge(prometheus.GaugeOpts{Name: "scheduled_unix", Help: "random help message"})
	prometheus.MustRegister(gauge)

	go func() {
		for t := range time.Tick(time.Second * 5) {
			gauge.Set(float64(t.Unix()))
		}
	}()

	//This section will start the HTTP server and expose
	//any metrics on the /metrics endpoint.
	http.Handle("/metrics", promhttp.Handler())
	log.Info("Beginning to serve on port 0.0.0.0:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
