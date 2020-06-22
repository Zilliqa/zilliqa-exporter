package main

import (
	"encoding/json"
	"genet_exporter/collector"
	"genet_exporter/logrusmiddleware"
	_ "github.com/joho/godotenv/autoload"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"path/filepath"
)

var options = &collector.Options{}
var listen string

var cmd = &cobra.Command{
	Use:   filepath.Base(os.Args[0]),
	Short: "zilliqa metric exporter",
	RunE: func(cmd *cobra.Command, args []string) error {
		return serve(listen)
	},
}

func main() {
	log.SetOutput(os.Stderr)
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	log.SetLevel(log.InfoLevel)
	if os.Getenv("DEBUG") != "" {
		log.SetLevel(log.DebugLevel)
	}
	options.BindFlags(cmd.Flags())
	cmd.Flags().StringVarP(&listen, "listen", "l", "127.0.0.1:8080", "listen address of exporter")
	_ = cmd.Execute()

}

func serve(listen string) error {
	//Create a new instance of the foo collector and
	//register it with the prometheus client.
	constants := collector.GetConstants(options)
	constants.Register(prometheus.DefaultRegisterer)

	optionsJson, _ := json.Marshal(options)
	log.WithField("options", string(optionsJson)).Debug("run with options")
	constantJson, _ := json.Marshal(constants)
	log.WithField("constants", string(constantJson)).Debug("got constants")

	if !options.NotCollectAPI {
		prometheus.MustRegister(collector.NewAPICollector(constants))
	}
	if !options.NotCollectAdmin {
		prometheus.MustRegister(collector.NewAdminCollector(constants))
	}
	if !options.NotCollectProcessInfo {
		prometheus.MustRegister(collector.NewProcessInfoCollector(constants))
	}

	l := logrusmiddleware.Middleware{
		Name:   "example",
		Logger: log.StandardLogger(),
	}
	http.Handle("/metrics", l.Handler(promhttp.Handler(), "metrics"))
	log.Infof("Beginning to serve on port %s", listen)
	return http.ListenAndServe(listen, nil)
}
