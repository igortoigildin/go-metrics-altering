// Package memory accumulates updated metrics,
// which is being used and consumed by agent.

package memory

import (
	"errors"
	"math/rand/v2"
	"runtime"
	"sync"
	"time"

	config "github.com/igortoigildin/go-metrics-altering/config/agent"
	"github.com/igortoigildin/go-metrics-altering/internal/models"
	"github.com/igortoigildin/go-metrics-altering/pkg/logger"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"go.uber.org/zap"
)

var (
	ErrConnectionFailed = errors.New("connection failed")
)

// MemoryStats provides temporary storage for collected metrics.
type MemoryStats struct {
	GaugeMetrics  map[string]float64
	CounterMetric int
	RunTimeMem    *runtime.MemStats
	rwm           sync.RWMutex
}

// NewMemoryStats is constructor for MemoryStats.
func NewMemoryStats() *MemoryStats {
	return &MemoryStats{
		GaugeMetrics: make(map[string]float64),
		RunTimeMem:   &runtime.MemStats{},
	}
}

// ReadMetrics collects runtime metrics from memory and sends it to agent via metricsChan.
func (m *MemoryStats) ReadMetrics(cfg *config.ConfigAgent, metricsChan chan models.Metrics) {
	for {
		time.Sleep(cfg.PauseDuration)
		m.rwm.Lock()
		for name, value := range m.GaugeMetrics {
			metric := models.GaugeConstructor(value, name)
			metricsChan <- metric
		}
		m.rwm.Unlock()
		metric := models.CounterConstructor(int64(m.CounterMetric))
		metricsChan <- metric
	}
}

// UpdateCPURAMStat collects and updates in MemoryStats CPU metrics.
func (m *MemoryStats) UpdateCPURAMStat(cfg *config.ConfigAgent) {
	if m.GaugeMetrics == nil {
		m.GaugeMetrics = make(map[string]float64, 3)
	}
	for {
		time.Sleep(cfg.PauseDuration)
		cpuNumber, err := cpu.Counts(true)
		if err != nil {
			logger.Log.Error("error while loading cpu Counts:", zap.Error(err))
		}
		v, err := mem.VirtualMemory()
		if err != nil {
			logger.Log.Error("error while loading virtualmemoryStat:", zap.Error(err))
		}
		m.rwm.Lock()
		m.GaugeMetrics["TotalMemory"] = float64(v.Total)
		m.GaugeMetrics["FreeMemory"] = float64(v.Free)
		m.GaugeMetrics["CPUutilization1"] = float64(cpuNumber)
		m.rwm.Unlock()
	}
}

// UpdateRunTimeStat collects and updates in MemoryStats runtime metrics.
func (m *MemoryStats) UpdateRunTimeStat(cfg *config.ConfigAgent) {
	for {
		time.Sleep(cfg.PauseDuration)
		runtime.ReadMemStats(m.RunTimeMem)
		m.rwm.Lock()
		m.GaugeMetrics["Alloc"] = float64(m.RunTimeMem.Alloc)
		m.GaugeMetrics["BuckHashSys"] = float64(m.RunTimeMem.BuckHashSys)
		m.GaugeMetrics["Frees"] = float64(m.RunTimeMem.Frees)
		m.GaugeMetrics["GCCPUFraction"] = float64(m.RunTimeMem.GCCPUFraction)
		m.GaugeMetrics["GCSys"] = float64(m.RunTimeMem.GCSys)
		m.GaugeMetrics["HeapAlloc"] = float64(m.RunTimeMem.HeapAlloc)
		m.GaugeMetrics["HeapIdle"] = float64(m.RunTimeMem.HeapIdle)
		m.GaugeMetrics["HeapInuse"] = float64(m.RunTimeMem.HeapInuse)
		m.GaugeMetrics["HeapObjects"] = float64(m.RunTimeMem.HeapObjects)
		m.GaugeMetrics["HeapReleased"] = float64(m.RunTimeMem.HeapReleased)
		m.GaugeMetrics["HeapSys"] = float64(m.RunTimeMem.HeapSys)
		m.GaugeMetrics["LastGC"] = float64(m.RunTimeMem.LastGC)
		m.GaugeMetrics["Lookups"] = float64(m.RunTimeMem.Lookups)
		m.GaugeMetrics["MCacheInuse"] = float64(m.RunTimeMem.MCacheInuse)
		m.GaugeMetrics["MCacheSys"] = float64(m.RunTimeMem.MCacheSys)
		m.GaugeMetrics["MSpanInuse"] = float64(m.RunTimeMem.MSpanInuse)
		m.GaugeMetrics["MSpanSys"] = float64(m.RunTimeMem.MSpanSys)
		m.GaugeMetrics["Mallocs"] = float64(m.RunTimeMem.Mallocs)
		m.GaugeMetrics["NextGC"] = float64(m.RunTimeMem.NextGC)
		m.GaugeMetrics["NumForcedGC"] = float64(m.RunTimeMem.NumForcedGC)
		m.GaugeMetrics["NumGC"] = float64(m.RunTimeMem.NumGC)
		m.GaugeMetrics["OtherSys"] = float64(m.RunTimeMem.OtherSys)
		m.GaugeMetrics["NextGC"] = float64(m.RunTimeMem.NextGC)
		m.GaugeMetrics["NumForcedGC"] = float64(m.RunTimeMem.NumForcedGC)
		m.GaugeMetrics["NumGC"] = float64(m.RunTimeMem.NumGC)
		m.GaugeMetrics["OtherSys"] = float64(m.RunTimeMem.OtherSys)
		m.GaugeMetrics["PauseTotalNs"] = float64(m.RunTimeMem.PauseTotalNs)
		m.GaugeMetrics["StackInuse"] = float64(m.RunTimeMem.StackInuse)
		m.GaugeMetrics["StackSys"] = float64(m.RunTimeMem.StackSys)
		m.GaugeMetrics["Sys"] = float64(m.RunTimeMem.StackSys)
		m.GaugeMetrics["TotalAlloc"] = float64(m.RunTimeMem.TotalAlloc)
		m.GaugeMetrics["RandomValue"] = rand.Float64()
		m.CounterMetric++
		m.rwm.Unlock()
	}
}
