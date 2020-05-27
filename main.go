package main

import _ "github.com/joho/godotenv/autoload"

import (
	"flag"
	"genet_exporter/logrusmiddleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
)

var options CollectorOptions

func main() {
	log.SetOutput(os.Stderr)
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	log.SetLevel(log.InfoLevel)
	if os.Getenv("DEBUG") != "" {
		log.SetLevel(log.DebugLevel)
	}

	options.BindFlags(flag.CommandLine)
	flag.Parse()

	//Create a new instance of the foo collector and
	//register it with the prometheus client.
	scheduledCollector := NewScheduledCollector(options)
	scheduledCollector.Init(prometheus.DefaultRegisterer)

	constLabels := prometheus.Labels{
		"type":         options.NodeType,
		"cluster_name": scheduledCollector.ClusterName,
		"network_name": scheduledCollector.NetworkName,
		"pod_name":     scheduledCollector.PodName,
		"public_ip":    scheduledCollector.PublicIP,
		"local_ip":     scheduledCollector.LocalIP,
	}

	prometheus.MustRegister(newInstantCollector(options, constLabels))
	if options.IsSameNS {
		prometheus.MustRegister(NewPsutilCollector(constLabels))
	}

	l := logrusmiddleware.Middleware{
		Name:   "example",
		Logger: log.StandardLogger(),
	}

	//This section will start the HTTP server and expose
	//any metrics on the /metrics endpoint.
	http.Handle("/metrics", l.Handler(promhttp.Handler(), "metrics"))
	log.Info("Beginning to serve on port 0.0.0.0:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
