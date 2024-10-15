package local

import (
	"sync"

	"github.com/igortoigildin/go-metrics-altering/internal/models"
)

type MetricAlgo interface {
	Update(metricType string, metricName string, metricValue any) error
	Get(metricType string, metricName string) (models.Metrics, error)
}

type counterRepo struct {
	Counter map[string]int64
	rm      sync.Mutex
}

func (c *counterRepo) Update(metricType string, metricName string, metricValue any) error {
	if c.Counter == nil {
		c.Counter = make(map[string]int64)
	}
	v, _ := metricValue.(int64)
	c.rm.Lock()
	c.Counter[metricName] += v
	c.rm.Unlock()
	return nil
}

func (c *counterRepo) Get(metricType string, metricName string) (models.Metrics, error) {
	var metric models.Metrics

	c.rm.Lock()
	d := c.Counter[metricName]
	c.rm.Unlock()
	metric.Delta = &d
	metric.MType = metricType

	return metric, nil
}

type gaugeRepo struct {
	Gauge map[string]float64
	rm    sync.Mutex
}

func (g *gaugeRepo) Update(metricType string, metricName string, metricValue any) error {
	if g.Gauge == nil {
		g.Gauge = make(map[string]float64)
	}
	v, _ := metricValue.(float64)
	g.rm.Lock()
	g.Gauge[metricName] = v
	g.rm.Unlock()
	return nil
}
func (g *gaugeRepo) Get(metricType string, metricName string) (models.Metrics, error) {
	var metric models.Metrics

	g.rm.Lock()
	v := g.Gauge[metricName]
	metric.Value = &v
	g.rm.Unlock()
	metric.MType = metricType

	return metric, nil
}
