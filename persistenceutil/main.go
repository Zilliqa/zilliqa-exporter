package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strconv"
)

var mainCmd = &cobra.Command{
	Use:   filepath.Base(os.Args[0]),
	Short: "Tools for zilliqa persistence",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var checkCmd = &cobra.Command{
	Use:   "check PATH ...",
	Short: "check if levelDB database is corrupted",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("files: %s", args)
		corrupted := 0
		for _, path := range args {
			err := checkCorruption(path)
			if err != nil {
				log.WithError(err).WithField("path", path).Error("check fail")
				corrupted = 1
			} else {
				log.WithField("path", path).Info("OK")
			}
		}
		os.Exit(corrupted)
	},
}

func main() {
	var logLevel string
	cobra.OnInitialize(func() { // init log
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
	})
	mainCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "log level")
	mainCmd.AddCommand(
		checkCmd,
	)
	_ = mainCmd.Execute()
}
