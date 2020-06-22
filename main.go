package main

import (
	"encoding/json"
	"fmt"
	_ "github.com/joho/godotenv/autoload"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/zilliqa/genet_exporter/collector"
	"github.com/zilliqa/genet_exporter/logrusmiddleware"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

var options = &collector.Options{}
var listen string
var logLevel string
var printVersion bool
var (
	version   = "dev"
	commit    = ""
	branch    = ""
	tag       = ""
	date      = ""
	buildInfo = ""
)

var cmd = &cobra.Command{
	Use:   filepath.Base(os.Args[0]),
	Short: "zilliqa metric exporter",
	RunE: func(cmd *cobra.Command, args []string) error {
		return serve(listen)
	},
}

func initlog() {
	log.SetOutput(os.Stderr)
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	debug, err := strconv.ParseBool(os.Getenv("DEBUG"))
	if err == nil && debug && log.DebugLevel > level {
		level = log.DebugLevel
	}
	log.SetLevel(level)
	if log.GetLevel() >= log.TraceLevel {
		log.SetReportCaller(true)
	}
	log.Debugf("Loglevel set to '%v'", log.GetLevel())
}

func main() {
	cmd.SilenceErrors = false
	cmd.SilenceUsage = true
	options.BindFlags(cmd.Flags())
	cmd.Flags().StringVarP(&listen, "listen", "l", "127.0.0.1:8080", "listen address of exporter")
	cmd.Flags().StringVar(&logLevel, "log-level", "info", "log level")
	cmd.Flags().BoolVarP(&printVersion, "version", "v", false, "print version info")
	cobra.OnInitialize(initlog)
	_ = cmd.Execute()

}

func serve(listen string) error {
	if printVersion {
		var isoDate string
		u, err := strconv.ParseInt(version, 10, 64)
		if err != nil {
			isoDate = date
		}
		isoDate = time.Unix(u, 0).Format("2006-01-02T15:04:05-07") // iso8601
		fmt.Printf(
			"Version(%s) Date(%s) Branch(%s) Tag(%s) Commit(%s) BuildInfo(%s)\n",
			version, isoDate, branch, tag, commit, buildInfo,
		)
		return nil
	}

	constants := collector.GetConstants(options)
	constants.Register(prometheus.DefaultRegisterer)

	log.WithFields(options.ToMap()).Info("run with options")
	constantJson, _ := json.Marshal(constants)
	var constantsMap map[string]interface{}
	_ = json.Unmarshal(constantJson, &constantsMap)
	log.WithFields(constantsMap).Info("got constants")

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
