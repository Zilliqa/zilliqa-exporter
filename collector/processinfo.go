package collector

import (
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	log "github.com/sirupsen/logrus"
	"github.com/zilliqa/zilliqa-exporter/utils"
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

	// TODO: move these to container info
	// gopsutil cpu.Times()
	nodeCPUUsageSeconds *prometheus.Desc
	nodeCPUCoresCount   *prometheus.Desc
	// /sys/fs/cgroup/cpuacct/cpuacct.usage / 1e9
	containerCPUUsageSeconds *prometheus.Desc
	// /sys/fs/cgroup/cpuacct/cpu.cfs_quota_us
	containerCPUCFSQuotaMicroseconds *prometheus.Desc
	// /sys/fs/cgroup/cpuacct/cpu.cfs_period_us
	// https://www.kernel.org/doc/Documentation/scheduler/sched-bwc.txt
	containerCPUCFSPeriodMicroseconds *prometheus.Desc
	containerCPUCoresLimitEquivalence *prometheus.Desc
	// from psutil process.Process.Times()
	processCPUUsageSeconds *prometheus.Desc

	// gopsutil mem.VirtualMemory()
	nodeMemUsageBytes *prometheus.Desc
	nodeMemTotalBytes *prometheus.Desc
	// no swap for k8s
	// /sys/fs/cgroup/memory/memory.usage_in_bytes
	containerMemUsageBytes *prometheus.Desc
	// /sys/fs/cgroup/memory/memory.limit_in_bytes or node capacity
	containerMemLimitBytes *prometheus.Desc

	processMemUsageBytes *prometheus.Desc

	// /run/zilliqa
	storageTotal *prometheus.Desc
	storageUsed  *prometheus.Desc
}

var processLabels = []string{"process_name", "pid", "cwd"}

func NewProcessInfoCollector(constants *Constants) *ProcessInfoCollector {
	commonLabels := constants.CommonLabels()
	processCommonLabels := append(commonLabels, processLabels...)
	return &ProcessInfoCollector{
		options:   constants.options,
		constants: constants,
		processRunning: prometheus.NewDesc(
			"zilliqa_process_running", "If zilliqa process is running",
			processCommonLabels, nil,
		),
		syncType: prometheus.NewDesc(
			"synctype", "Synctype from zilliqa commandline options",
			processCommonLabels, nil,
		),
		nodeType: prometheus.NewDesc(
			"nodetype", "Nodetype from zilliqa commandline options",
			append([]string{"text"}, processCommonLabels...), nil,
		),
		nodeIndex: prometheus.NewDesc(
			"nodeindex", "Nodeindex from zilliqa commandline options",
			processCommonLabels, nil,
		),

		uptime: prometheus.NewDesc(
			"node_uptime", "Uptime of zilliqa node",
			processCommonLabels, nil,
		),
		connectionCount: prometheus.NewDesc(
			"connection_count", "Connection count of zilliqa process",
			append([]string{"local_port", "status"}, processCommonLabels...), nil,
		),
		threadCount: prometheus.NewDesc(
			"thread_count", "Thread count of zilliqa process",
			processCommonLabels, nil,
		),
		fdCount: prometheus.NewDesc(
			"fd_count", "Opened files count of zilliqa process",
			processCommonLabels, nil,
		),

		nodeCPUUsageSeconds: prometheus.NewDesc(
			"node_cpu_usage_seconds", "cpu usage in nano seconds of the node",
			commonLabels, nil,
		),
		nodeCPUCoresCount: prometheus.NewDesc(
			"node_cpu_cores_count", "count of node logic CPU cores",
			commonLabels, nil,
		),
		containerCPUUsageSeconds: prometheus.NewDesc(
			"container_cpu_usage_seconds", "cpu usage in nano seconds of the container",
			commonLabels, nil,
		),
		containerCPUCFSQuotaMicroseconds: prometheus.NewDesc(
			"container_cpu_cfs_quota_microseconds", "cpu CFS quota in microseconds of the container",
			commonLabels, nil,
		),
		containerCPUCFSPeriodMicroseconds: prometheus.NewDesc(
			"container_cpu_cfs_period_microseconds", "cpu CFS period in microseconds of the container",
			commonLabels, nil,
		),
		containerCPUCoresLimitEquivalence: prometheus.NewDesc(
			"container_cpu_cores_limit_equivalence", "cpu cfs cores limit of the container, if no limit, set to physical limit",
			commonLabels, nil,
		),


		nodeMemUsageBytes: prometheus.NewDesc(
			"node_mem_usage_bytes", "memory usage in bytes of the node",
			commonLabels, nil,
		),
		nodeMemTotalBytes: prometheus.NewDesc(
			"node_mem_total_bytes", "memory total in bytes of the node",
			commonLabels, nil,
		),
		containerMemUsageBytes: prometheus.NewDesc(
			"container_mem_usage_bytes", "memory usage in bytes of the container",
			commonLabels, nil,
		),
		containerMemLimitBytes: prometheus.NewDesc(
			"container_mem_limit_bytes", "memory limit in bytes of the container",
			commonLabels, nil,
		),


		processCPUUsageSeconds: prometheus.NewDesc(
			"process_cpu_usage_seconds", "cpu usage in nano seconds of process",
			processCommonLabels, nil,
		),
		processMemUsageBytes: prometheus.NewDesc(
			"process_mem_usage_bytes", "memory usage in bytes of process",
			processCommonLabels, nil,
		),


		storageTotal: prometheus.NewDesc(
			"storage_total", "Total capacity of zilliqa persistence storage",
			processCommonLabels, nil,
		),
		storageUsed: prometheus.NewDesc(
			"storage_used", "Used space of zilliqa persistence storage",
			processCommonLabels, nil,
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

	ch <- c.nodeCPUUsageSeconds
	ch <- c.nodeCPUCoresCount
	ch <- c.containerCPUUsageSeconds
	ch <- c.containerCPUCFSQuotaMicroseconds
	ch <- c.containerCPUCFSPeriodMicroseconds
	ch <- c.processCPUUsageSeconds
	ch <- c.containerCPUCoresLimitEquivalence

	ch <- c.nodeMemUsageBytes
	ch <- c.nodeMemUsageBytes
	ch <- c.containerMemUsageBytes
	ch <- c.containerMemLimitBytes
	ch <- c.processMemUsageBytes

	// /run/zilliqa
	ch <- c.storageTotal
	ch <- c.storageUsed
}

func (c *ProcessInfoCollector) Collect(ch chan<- prometheus.Metric) {
	log.Debug("start collecting process info")
	process := GetZilliqaMainProcess(c.constants)
	processd := GetZilliqadProcess()
	if process == nil {
		log.Error("no running zilliqa process found")
		ch <- prometheus.MustNewConstMetric(c.processRunning, prometheus.GaugeValue, 0, append(c.constants.CommonLabelValues(), "", "0", "")...)
		return
	}
	pid := process.Pid
	cwd, _ := process.Cwd()
	commonValues := c.constants.CommonLabelValues()
	name, _ := process.Name()
	labels := append(commonValues, name, strconv.Itoa(int(pid)), cwd)
	ch <- prometheus.MustNewConstMetric(c.processRunning, prometheus.GaugeValue, 1, labels...)

	wg := sync.WaitGroup{}

	// synctype
	wg.Add(1)
	go func() {
		defer wg.Done()
		cmdline, _ := process.CmdlineSlice()
		if processd != nil {
			log.Debug("use zilliqad's cmdline for process info")
			cmdline, _ = processd.CmdlineSlice()
		}
		log.WithField("cmdline", cmdline).Debug("found cmdline of zilliqa process")
		syncType, err := GetSyncTypeFromCmdline(cmdline)
		if err != nil {
			log.WithError(err).Error("fail to get sync type")
		} else {
			log.WithField("syncType", syncType).Debug("found sync type")
			ch <- prometheus.MustNewConstMetric(c.syncType, prometheus.GaugeValue, float64(syncType), labels...)
		}

		nt := GetNodeTypeFromCmdline(cmdline)
		nodeType := NodeTypeFromString(nt)
		if nodeType == UnknownNodeType {
			log.Errorf("fail to get node type, type %s unknown", nodeType)
		} else {
			log.WithField("nodeType", nodeType).Debug("found node type")
			ch <- prometheus.MustNewConstMetric(
				c.nodeType, prometheus.GaugeValue, float64(nodeType),
				append([]string{nodeType.String()}, labels...)...,
			)
		}

		nodeIndex, err := GetNodeIndexFromCmdline(cmdline)
		if err != nil {
			log.WithError(err).Error("fail to get node index")
		} else {
			log.WithField("nodeIndex", nodeIndex).Debug("found node index")
			ch <- prometheus.MustNewConstMetric(c.nodeIndex, prometheus.GaugeValue, float64(nodeIndex), labels...)
		}
	}()

	// uptime
	wg.Add(1)
	go func() {
		defer wg.Done()
		// milliseconds since the epoch, in UTC
		created, _ := process.CreateTime()
		//nanoNow := time.Now().UnixNano()
		//uptime := time.Duration(nanoNow)*time.Nanosecond - time.Duration(created)*time.Millisecond
		ch <- prometheus.MustNewConstMetric(c.uptime, prometheus.GaugeValue, float64(created), labels...)
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
		} else {
			log.WithError(err).Error("error while getting threadCount")
		}
		fds, err := process.NumFDs()
		if err == nil {
			ch <- prometheus.MustNewConstMetric(c.fdCount, prometheus.GaugeValue, float64(fds), labels...)
		} else {
			log.WithError(err).Error("error while getting fdCount")
		}

		// cpu of zilliqa main process
		t, err := process.Times()
		if err == nil {
			zCPUSecs := t.User + t.System + t.Nice + t.Iowait + t.Irq +
				t.Softirq + t.Steal
			ch <- prometheus.MustNewConstMetric(c.processCPUUsageSeconds, prometheus.GaugeValue, zCPUSecs, labels...)
		} else {
			log.WithError(err).Error("error while getting processCPUUsageSeconds")
		}
		// mem of zilliqa main process
		zMem, err := process.MemoryInfo()
		if err == nil {
			ch <- prometheus.MustNewConstMetric(c.processMemUsageBytes, prometheus.GaugeValue, float64(zMem.RSS), labels...)
		} else {
			log.WithError(err).Error("error while getting process mem info")
		}

		// storage of zilliqa working dir (/run/zilliqa)
		storageStats, err := disk.Usage(cwd)
		if err == nil {
			ch <- prometheus.MustNewConstMetric(c.storageTotal, prometheus.GaugeValue, float64(storageStats.Total), labels...)
			ch <- prometheus.MustNewConstMetric(c.storageUsed, prometheus.GaugeValue, float64(storageStats.Used), labels...)
		} else {
			log.WithError(err).Error("error while getting storageStats")
		}
	}()

	// container and node info
	wg.Add(1)
	go func() {
		defer wg.Done()
		// cpu of node
		nodeTimes, err := cpu.Times(false)
		if err == nil {
			t := nodeTimes[0]
			nodeCPUSecs := t.User + t.System + t.Nice + t.Iowait + t.Irq +
				t.Softirq + t.Steal
			ch <- prometheus.MustNewConstMetric(c.nodeCPUUsageSeconds, prometheus.GaugeValue, nodeCPUSecs, commonValues...)
		} else {
			log.WithError(err).Error("fail to get node CPU info")
		}
		cpuCount, err := cpu.Counts(true)
		if err == nil {
			ch <- prometheus.MustNewConstMetric(c.nodeCPUCoresCount, prometheus.GaugeValue, float64(cpuCount), commonValues...)
		} else {
			log.WithError(err).Error("error while getting nodeCPUCoresCount")
		}
		acctPath := "/sys/fs/cgroup/cpuacct/"
		cpuPath := "/sys/fs/cgroup/cpu/"

		// cpu of entire container (current cgroup)
		if utils.PathIsDir(acctPath) || utils.PathIsDir(cpuPath) {
			usageNano, err := utils.ReadFloat64("/sys/fs/cgroup/cpuacct/cpuacct.usage")
			usageSecs := usageNano / float64(time.Second)
			if err == nil {
				ch <- prometheus.MustNewConstMetric(c.containerCPUUsageSeconds, prometheus.GaugeValue, usageSecs, commonValues...)
			} else {
				log.WithError(err).Error("error while getting containerCPUUsageSeconds")
			}
			quota, err := utils.ReadFloat64(utils.FindFile([]string{acctPath, cpuPath}, []string{"cpu.cfs_quota_us"}))
			if err == nil {
				ch <- prometheus.MustNewConstMetric(c.containerCPUCFSQuotaMicroseconds, prometheus.GaugeValue, quota, commonValues...)
			} else {
				log.WithError(err).Error("error while getting containerCPUCFSQuotaMicroseconds")
			}
			period, err := utils.ReadFloat64(utils.FindFile([]string{acctPath, cpuPath}, []string{"cpu.cfs_period_us"}))
			if err == nil {
				ch <- prometheus.MustNewConstMetric(c.containerCPUCFSPeriodMicroseconds, prometheus.GaugeValue, period, commonValues...)
			} else {
				log.WithError(err).Error("error while getting containerCPUCFSPeriodMicroseconds")
			}
			coresLimit := float64(cpuCount)
			if quota > 0 {
				coresLimit = quota / period
			}
			ch <- prometheus.MustNewConstMetric(c.containerCPUCoresLimitEquivalence, prometheus.GaugeValue, coresLimit, commonValues...)
		} else {
			log.Warn("/sys/fs/cgroup/cpuacct or /sys/fs/cgroup/cpu not found, skip collecting container cpu info")
		}

		// mem info of node
		nodeMem, err := mem.VirtualMemory()
		if err == nil {
			ch <- prometheus.MustNewConstMetric(c.nodeMemUsageBytes, prometheus.GaugeValue, float64(nodeMem.Used), commonValues...)
			ch <- prometheus.MustNewConstMetric(c.nodeMemTotalBytes, prometheus.GaugeValue, float64(nodeMem.Total), commonValues...)

			// mem info of container
			if utils.PathIsDir("/sys/fs/cgroup/memory/") {
				usage, err := utils.ReadFloat64("/sys/fs/cgroup/memory/memory.usage_in_bytes")
				if err == nil {
					ch <- prometheus.MustNewConstMetric(c.containerMemUsageBytes, prometheus.GaugeValue, usage, commonValues...)
				} else {
					log.WithError(err).Error("error while getting containerMemUsageBytes")
				}
				limit, err := utils.ReadFloat64("/sys/fs/cgroup/memory/memory.limit_in_bytes")
				if err == nil {
					if limit == 0 || limit > float64(nodeMem.Total) {
						limit = float64(nodeMem.Total)
					}
					ch <- prometheus.MustNewConstMetric(c.containerMemLimitBytes, prometheus.GaugeValue, limit, commonValues...)
				} else {
					log.WithError(err).Error("error while getting containerMemLimitBytes")
				}
			} else {
				log.Warn("/sys/fs/cgroup/memory path not found, skip collecting container mem info")
			}
		} else {
			log.WithError(err).Error("fail to get node mem info")
		}
	}()
	wg.Wait()
	log.Debug("end collecting process info")
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
	value := GetParamValueFromCmdline(cmdline, "--synctype", "-s")
	if value == "" {
		return 0, errors.New("not found")
	}
	return strconv.Atoi(value)
}

func GetNodeTypeFromCmdline(cmdline []string) string {
	return GetParamValueFromCmdline(cmdline, "--nodetype", "-n")
}

func GetNodeIndexFromCmdline(cmdline []string) (int, error) {
	value := GetParamValueFromCmdline(cmdline, "--nodeindex", "-x")
	if value == "" {
		return 0, errors.New("not found")
	}
	return strconv.Atoi(value)
}

func GetPortFromCmdline(cmdline []string) (int, error) {
	value := GetParamValueFromCmdline(cmdline, "--port", "-p")
	if value == "" {
		return 0, errors.New("not found")
	}
	return strconv.Atoi(value)
}

func GetParamValueFromCmdline(cmdline []string, param ...string) string {
	for i, arg := range cmdline {
		for _, p := range param {
			if strings.HasPrefix(arg, p) {
				if strings.Contains(arg, "=") {
					splits := strings.Split(arg, "=")
					// --opt value
					if len(splits) < 2 {
						return cmdline[i+1]
					}
					// --opt=value
					return splits[1]
				}
				return cmdline[i+1]
			}
		}
	}
	return ""
}
