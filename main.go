package main

import (
	"encoding/json"
	"fmt"
	_ "github.com/joho/godotenv/autoload"
	"github.com/meatballhat/negroni-logrus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/urfave/negroni"
	"github.com/zilliqa/zilliqa-exporter/collector"
	"net/http"
	"net/http/pprof"
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

func versionOutput() string {
	var isoDate string
	u, err := strconv.ParseInt(date, 10, 64)
	if err != nil {
		isoDate = date
	} else {
		isoDate = time.Unix(u, 0).Format("2006-01-02T15:04:05-07") // iso8601
	}
	return fmt.Sprintf(
		"Version(%s) Date(%s) Branch(%s) Tag(%s) Commit(%s) BuildInfo(%s)",
		version, isoDate, branch, tag, commit, buildInfo,
	)
}

var cmd = &cobra.Command{
	Use:   filepath.Base(os.Args[0]),
	Short: "zilliqa metric exporter",
	Long:  versionOutput(),
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
	cmd.Flags().StringVarP(&listen, "listen", "l", "0.0.0.0:8080", "listen address of exporter")
	cmd.Flags().StringVar(&logLevel, "log-level", "info", "log level")
	cmd.Flags().BoolVarP(&printVersion, "version", "v", false, "print version info")
	cobra.OnInitialize(initlog)
	_ = cmd.Execute()

}

func serve(listen string) error {
	if printVersion {
		fmt.Println(versionOutput())
		return nil
	}

	constants := collector.NewConstants(options)
	go constants.StartWatch()
	defer constants.StopWatch()
	prometheus.MustRegister(constants)

	versionInfo := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "zilliqa_exporter_version_info",
		Help: "the version information of zilliqa exporter",
		ConstLabels: prometheus.Labels{
			"version":      version,
			"commit":       commit,
			"branch":       branch,
			"tag":          tag,
			"date":         date,
			"buildInfo":    buildInfo,
			"type":         constants.NodeType().String(),
			"cluster_name": constants.ClusterName,
			"network_name": constants.NetworkName,
			"pod_name":     constants.PodName,
		},
	})
	prometheus.MustRegister(versionInfo)
	versionInfo.Set(1)

	log.WithFields(options.ToMap()).Info("run with options")
	constantJson, _ := json.Marshal(constants)
	var constantsMap map[string]interface{}
	_ = json.Unmarshal(constantJson, &constantsMap)
	log.WithFields(constantsMap).Info("got constants")

	if !options.NotCollectAPI {
		prometheus.MustRegister(collector.NewAPICollector(constants))
	} else {
		log.Info("Not collecting info from API server")
	}
	if !options.NotCollectAdmin {
		prometheus.MustRegister(collector.NewAdminCollector(constants))
	} else {
		log.Info("Not collecting info from Admin(status) server")
	}
	if !options.NotCollectProcessInfo {
		prometheus.MustRegister(collector.NewProcessInfoCollector(constants))
	} else {
		log.Info("Not collecting info from Zilliqa Process")
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/panic", func(w http.ResponseWriter, req *http.Request) {
		panic("panic test")
	})
	// bind pprof
	{
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	n := negroni.New()
	recovery := &negroni.Recovery{
		Logger:     log.StandardLogger(),
		PrintStack: log.GetLevel() >= log.DebugLevel,
		StackAll:   false,
		StackSize:  1024 * 8,
		Formatter:  &negroni.HTMLPanicFormatter{},
	}
	n.Use(negronilogrus.NewMiddlewareFromLogger(log.StandardLogger(), "req"))
	n.Use(recovery)
	n.UseHandler(mux)

	log.Info("Zilliqa Exporter")
	log.Info(versionOutput())
	log.Infof("Beginning to serve on port %s", listen)
	return http.ListenAndServe(listen, n)
}
