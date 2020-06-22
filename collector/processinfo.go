package collector

import (
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/disk"
	"github.com/zilliqa/genet_exporter/utils"

	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ProcessInfoCollector struct {
	options   *Options
	constants *Constants

	// os related, from psutil
	processRunning *prometheus.Desc
	syncType       *prometheus.Desc
	nodeType       *prometheus.Desc
	nodeIndex      *prometheus.Desc

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

func NewProcessInfoCollector(constants *Constants) *ProcessInfoCollector {
	constLabels := constants.ConstLabels()
	return &ProcessInfoCollector{
		options:   constants.options,
		constants: constants,
		processRunning: prometheus.NewDesc(
			"zilliqa_process_running", "If zilliqa process is running",
			processLabels, constLabels,
		),
		syncType: prometheus.NewDesc(
			"synctype", "Synctype from zilliqa commandline options",
			processLabels, constLabels,
		),
		nodeType: prometheus.NewDesc(
			"nodetype", "Nodetype from zilliqa commandline options",
			append([]string{"text"}, processLabels...), constLabels,
		),
		nodeIndex: prometheus.NewDesc(
			"nodeindex", "Nodeindex from zilliqa commandline options",
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

func (c *ProcessInfoCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.processRunning
	ch <- c.syncType
	ch <- c.nodeType
	ch <- c.nodeIndex

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

func (c *ProcessInfoCollector) Collect(ch chan<- prometheus.Metric) {
	process := utils.GetZilliqaMainProcess()
	if process == nil {
		log.Error("no running zilliqa process found")
		ch <- prometheus.MustNewConstMetric(c.processRunning, prometheus.GaugeValue, 0, "0", "")
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
		syncType, err := GetSyncTypeFromCmdline(cmdline)
		if err != nil {
			log.WithError(err).Error("fail to get sync type")
		} else {
			ch <- prometheus.MustNewConstMetric(c.syncType, prometheus.GaugeValue, float64(syncType), labels...)
		}

		nt := GetNodeTypeFromCmdline(cmdline)
		nodeType := NodeTypeFromString(nt)
		if nodeType == UnknownNodeType {
			log.Errorf("fail to get node type, type %s unknown", nodeType)
		} else {
			ch <- prometheus.MustNewConstMetric(
				c.nodeType, prometheus.GaugeValue, float64(nodeType),
				append([]string{nodeType.String()}, labels...)...,
			)
		}

		nodeIndex, err := GetNodeIndexFromCmdline(cmdline)
		if err != nil {
			log.WithError(err).Error("fail to get sync type")
		} else {
			ch <- prometheus.MustNewConstMetric(c.syncType, prometheus.GaugeValue, float64(nodeIndex), labels...)
		}
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
			if ct.Port <= 0 {
				continue
			}
			ch <- prometheus.MustNewConstMetric(
				c.connectionCount, prometheus.GaugeValue, count,
				append([]string{strconv.Itoa(int(ct.Port)), ct.Status}, labels...)...,
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

//# synctype:
//# 0(default) for no
//# 1 for new
//# 2 for normal
//# 3 for ds
//# 4 for lookup
//# 5 for node recovery
//# 6 for new lookup
//# 7 for ds guard node sync
//# 8 for offline validation of DB

type SyncType int

const (
	DefaultSyncType SyncType = iota
	NewSyncType
	NormalSyncType
	DSSyncType
	LookupSyncType
	RecoverySyncType
	NewLookupSyncType
	DSGuardNodeSyncSyncType
	OfflineDBValidationSyncType
)

func GetSyncTypeFromCmdline(cmdline []string) (int, error) {
	value := GetParamValueFromCmdline(cmdline, "--synctype")
	if value == "" {
		return 0, errors.New("not found")
	}
	return strconv.Atoi(value)
}

func GetNodeTypeFromCmdline(cmdline []string) string {
	return GetParamValueFromCmdline(cmdline, "--nodetype")
}

func GetNodeIndexFromCmdline(cmdline []string) (int, error) {
	value := GetParamValueFromCmdline(cmdline, "--nodeindex")
	if value == "" {
		return 0, errors.New("not found")
	}
	return strconv.Atoi(value)
}

func GetParamValueFromCmdline(cmdline []string, param string) string {
	for i, arg := range cmdline {
		if strings.HasPrefix(arg, param) {
			if strings.Contains(arg, "=") {
				splits := strings.Split(arg, "=")
				// --opt value
				if len(splits) < 2 {
					return cmdline[i+1]
				}
				// --opt=value
				return splits[1]
			}
		}
	}
	return ""
}
