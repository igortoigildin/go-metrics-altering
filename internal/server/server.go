package server

type MemStorage struct {
	Gauge     map[string]float64
	Counter   map[string]int64
}

func InitStorage() (*MemStorage) {
	m := new(MemStorage)
	m.Counter  = make(map[string]int64)
	m.Gauge = make(map[string]float64)
	return m
}

func (m *MemStorage) UpdateGaugeMetric(metricName string, metricValue float64) {
	m.Gauge[metricName] = metricValue
}

func (m *MemStorage) UpdateCounterMetric(metricName string, metricValue int64) {
	m.Counter[metricName] += metricValue
}

func (m *MemStorage) GetGaugeMetricFromMemory(metricName string) float64 {
	return m.Gauge[metricName]
}

func (m *MemStorage) GetCountMetricFromMemory(metricName string) int64 {
	return m.Counter[metricName]
}

func (m *MemStorage) CheckIfGaugeMetricPresent(metricName string) bool {
	_, ok := m.Gauge[metricName]
	return ok
}

func (m *MemStorage) CheckIfCountMetricPresent(metricName string) bool {
	_, ok := m.Counter[metricName]
	return ok
}

func ConvertToSingleMap(a map[string]float64, b map[string]int64) map[string]interface{} {
	c := make(map[string]interface{})
	for i, v := range a {
		c[i] = v
	}
	for j, l := range b {
		c[j] = l
	}
	return c
}

