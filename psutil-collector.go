package main

import (
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/disk"

	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"sync"
	"time"
)

type PsutilCollector struct {
	// os related, from psutil
	up              *prometheus.Desc
	syncType        *prometheus.Desc
	uptime          *prometheus.Desc
	connectionCount *prometheus.Desc
	threadCount     *prometheus.Desc
	fdCount         *prometheus.Desc

	//cpuPercent *prometheus.Desc
	//memPercent *prometheus.Desc

	// /run/zilliqa
	storageTotal *prometheus.Desc
	storageUsed  *prometheus.Desc
}

var processLabels = []string{"pid", "cwd"}

func NewPsutilCollector(constLabels prometheus.Labels) *PsutilCollector {
	return &PsutilCollector{
		up: prometheus.NewDesc(
			"up", "If zilliqa process is running",
			processLabels, constLabels,
		),
		syncType: prometheus.NewDesc(
			"synctype", "Synctype from zilliqa commandline option",
			processLabels, constLabels,
		),
		uptime: prometheus.NewDesc(
			"node_uptime", "Uptime of zilliqa node",
			processLabels, constLabels,
		),
		connectionCount: prometheus.NewDesc(
			"connection_count", "Connection count of zilliqa process",
			append([]string{"local_port", "status"}, processLabels...), constLabels,
		),
		threadCount: prometheus.NewDesc(
			"thread_count", "Thread count of zilliqa process",
			processLabels, constLabels,
		),
		fdCount: prometheus.NewDesc(
			"fd_count", "Opened files count of zilliqa process",
			processLabels, constLabels,
		),
		//cpuPercent: prometheus.NewDesc(
		//	"cpu_percent", "CPU usage percent of zilliqa process",
		//	processLabels, constLabels,
		//),
		//memPercent: prometheus.NewDesc(
		//	"mem_percent", "Memory usage percent of zilliqa process",
		//	processLabels, constLabels,
		//),
		storageTotal: prometheus.NewDesc(
			"storage_total", "Total capacity of zilliqa persistence storage",
			processLabels, constLabels,
		),
		storageUsed: prometheus.NewDesc(
			"storage_used", "Used space of zilliqa persistence storage",
			processLabels, constLabels,
		),
	}
}

func (c *PsutilCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.up
	ch <- c.syncType
	ch <- c.uptime
	ch <- c.connectionCount
	ch <- c.threadCount
	ch <- c.fdCount

	//ch <- c.cpuPercent
	//ch <- c.memPercent

	// /run/zilliqa
	ch <- c.storageTotal
	ch <- c.storageUsed
}

func getSyncType(cmdline []string) (syncType int, err error) {
	for i, arg := range cmdline {
		if strings.HasPrefix(arg, "--synctype") {
			if strings.Contains(arg, "=") {
				splits := strings.Split(arg, "=")
				if len(splits) < 2 {
					if len(cmdline) < i {
						err = errors.New("not found")
						return
					}
					syncType, err = strconv.Atoi(cmdline[i+1])
					return
				}
				syncType, err = strconv.Atoi(splits[1])
				return
			}
		}
	}
	return
}

func (c *PsutilCollector) Collect(ch chan<- prometheus.Metric) {
	process := getZilliqaMainProcess()
	if process == nil {
		log.Error("no running zilliqa process found")
		ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, 0, "0", "")
		return
	}
	pid := process.Pid
	cwd, _ := process.Cwd()
	labels := []string{strconv.Itoa(int(pid)), cwd}

	wg := sync.WaitGroup{}

	// synctype
	wg.Add(1)
	go func() {
		defer wg.Done()
		cmdline, _ := process.CmdlineSlice()
		syncType, err := getSyncType(cmdline)
		if err != nil {
			log.WithError(err).Error("fail to get sync type")
			return
		}
		ch <- prometheus.MustNewConstMetric(c.syncType, prometheus.GaugeValue, float64(syncType), labels...)
	}()

	// uptime
	wg.Add(1)
	go func() {
		defer wg.Done()
		created, _ := process.CreateTime()
		nanoNow := time.Now().UnixNano()
		uptime := time.Duration(nanoNow)*time.Nanosecond - time.Duration(created)*time.Millisecond
		ch <- prometheus.MustNewConstMetric(c.uptime, prometheus.GaugeValue, float64(uptime), labels...)
	}()

	// connections
	type connType struct {
		Port   uint32
		Status string
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		connections, err := process.Connections()
		if err != nil {
			return
		}
		counts := make(map[connType]float64)
		for _, conn := range connections {
			typ := connType{conn.Laddr.Port, conn.Status}
			counts[typ] += 1
		}
		for ct, count := range counts {
			ch <- prometheus.MustNewConstMetric(
				c.connectionCount, prometheus.GaugeValue, count,
				append([]string{strconv.Itoa(int(ct.Port)), ct.Status}, labels...)...
			)
		}
	}()

	// others
	wg.Add(1)
	go func() {
		defer wg.Done()
		threads, err := process.NumThreads()
		if err == nil {
			ch <- prometheus.MustNewConstMetric(c.threadCount, prometheus.GaugeValue, float64(threads), labels...)
		}
		fds, err := process.NumFDs()
		if err == nil {
			ch <- prometheus.MustNewConstMetric(c.fdCount, prometheus.GaugeValue, float64(fds), labels...)
		}

		//ch <- prometheus.MustNewConstMetric(c.cpuPercent, prometheus.GaugeValue, 0, labels...)
		//ch <- prometheus.MustNewConstMetric(c.memPercent, prometheus.GaugeValue, 0, labels...)

		// /run/zilliqa
		storageStats, _ := disk.Usage(cwd)
		ch <- prometheus.MustNewConstMetric(c.storageTotal, prometheus.GaugeValue, float64(storageStats.Total), labels...)
		ch <- prometheus.MustNewConstMetric(c.storageUsed, prometheus.GaugeValue, float64(storageStats.Used), labels...)
	}()
	wg.Wait()
}
