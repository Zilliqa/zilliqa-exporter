package collector

import (
	"context"
	"encoding/json"
	"github.com/Zilliqa/gozilliqa-sdk/provider"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

// some
type ScheduledCollector struct {
	options   *Options
	constants *Constants

	// props
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// TODO: persistence-check
	UDStateSize    prometheus.Gauge
	UDStateEntries prometheus.Gauge
}

func NewScheduledCollector(option *Options, constants *Constants) *ScheduledCollector {
	c := &ScheduledCollector{options: option, constants: constants}
	return c
}

func (s *ScheduledCollector) Init(register prometheus.Registerer) {
	constLabels := s.constants.ConstLabels()
	s.ctx, s.cancel = context.WithCancel(context.Background())
	if s.options.IsMainNet {
		s.UDStateSize = prometheus.NewGauge(prometheus.GaugeOpts{
			Name:        "ud_state_size",
			Help:        "State data size of unstoppable domain contract",
			ConstLabels: constLabels,
		})
		s.UDStateEntries = prometheus.NewGauge(prometheus.GaugeOpts{
			Name:        "ud_state_entries",
			Help:        "State entries count of unstoppable domain contract",
			ConstLabels: constLabels,
		})
		register.MustRegister(s.UDStateSize, s.UDStateEntries)
	}
}

func (s *ScheduledCollector) GetUDContractStateSizeRecords() (int, int, error) {
	var address = "9611c53BE6d1b32058b2747bdeCECed7e1216793"
	api := provider.NewProvider(s.options.APIEndpoint())
	resp, err := api.GetSmartContractState(address)
	if err != nil {
		return 0, 0, err
	}
	js, err := json.Marshal(resp.Result)
	if err != nil {
		return 0, 0, err
	}
	size := len(js)
	var result map[string]json.RawMessage
	err = json.Unmarshal(js, &result)
	if err != nil {
		return 0, 0, err
	}
	records, _ := result["records"]
	var entries map[string]interface{}
	err = json.Unmarshal(records, &entries)
	if err != nil {
		return 0, 0, err
	}
	return size, len(entries), nil
}

func (s *ScheduledCollector) CollectGetUDContractStateSizeRecords() {
	size, records, err := s.GetUDContractStateSizeRecords()
	if err != nil {
		log.WithError(err).Error("fail to get UD Contract state")
	}
	s.UDStateSize.Set(float64(size))
	s.UDStateEntries.Set(float64(records))
}

type ScheduleTask struct {
	fun      func()
	interval time.Duration
}

func (s *ScheduledCollector) Start() {
	var schedules []ScheduleTask

	if s.options.IsMainNet {
		schedules = append(schedules, ScheduleTask{s.CollectGetUDContractStateSizeRecords, 1 * time.Hour})
	}

	for _, t := range schedules {
		go func() {
			s.wg.Add(1)
			defer s.wg.Done()
			for {
				select {
				case <-time.After(t.interval):
					t.fun()
				case <-s.ctx.Done():
					return
				}
			}
		}()
	}
}

func (s *ScheduledCollector) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
}
