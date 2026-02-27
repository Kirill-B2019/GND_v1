// | KB @CerbeRus - Nexus Invest Team
package monitoring

import (
	"fmt"
	"math"
	"os"
	"sync"
	"time"
)

// MetricType определяет тип метрики
type MetricType string

const (
	MetricGauge   MetricType = "gauge"   // Текущее значение (например, TPS, CPU)
	MetricCounter MetricType = "counter" // Счетчик (например, число блоков, транзакций)
	MetricHist    MetricType = "hist"    // Гистограмма (например, задержки)
)

// Metric описывает структуру метрики
type Metric struct {
	Name        string
	Type        MetricType
	Value       float64
	Labels      map[string]string
	LastUpdated time.Time
	Histogram   []float64 // Для гистограмм: значения
}

// MetricsRegistry управляет метриками блокчейна
type MetricsRegistry struct {
	metrics map[string]*Metric
	mutex   sync.RWMutex
	logFile *os.File
}

// NewMetricsRegistry создает новый реестр метрик и файл логирования (если путь не пустой)
func NewMetricsRegistry(logFilePath string) (*MetricsRegistry, error) {
	var file *os.File
	var err error
	if logFilePath != "" {
		file, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, err
		}
	}
	return &MetricsRegistry{
		metrics: make(map[string]*Metric),
		logFile: file,
	}, nil
}

// SetGauge устанавливает значение gauge-метрики (например, TPS, CPU)
func (mr *MetricsRegistry) SetGauge(name string, value float64, labels map[string]string) {
	mr.mutex.Lock()
	defer mr.mutex.Unlock()
	key := metricKey(name, labels)
	metric, ok := mr.metrics[key]
	if !ok {
		metric = &Metric{Name: name, Type: MetricGauge, Labels: labels}
		mr.metrics[key] = metric
	}
	metric.Value = value
	metric.LastUpdated = time.Now()
	mr.logMetric(metric)
}

// IncCounter увеличивает счетчик на 1 (или на delta)
func (mr *MetricsRegistry) IncCounter(name string, delta float64, labels map[string]string) {
	mr.mutex.Lock()
	defer mr.mutex.Unlock()
	key := metricKey(name, labels)
	metric, ok := mr.metrics[key]
	if !ok {
		metric = &Metric{Name: name, Type: MetricCounter, Labels: labels}
		mr.metrics[key] = metric
	}
	metric.Value += delta
	metric.LastUpdated = time.Now()
	mr.logMetric(metric)
}

// ObserveHist добавляет значение в гистограмму (например, задержка блока)
func (mr *MetricsRegistry) ObserveHist(name string, value float64, labels map[string]string) {
	mr.mutex.Lock()
	defer mr.mutex.Unlock()
	key := metricKey(name, labels)
	metric, ok := mr.metrics[key]
	if !ok {
		metric = &Metric{Name: name, Type: MetricHist, Labels: labels}
		mr.metrics[key] = metric
	}
	metric.Histogram = append(metric.Histogram, value)
	metric.LastUpdated = time.Now()
	mr.logMetric(metric)
}

// GetMetric возвращает метрику по имени и лейблам
func (mr *MetricsRegistry) GetMetric(name string, labels map[string]string) *Metric {
	mr.mutex.RLock()
	defer mr.mutex.RUnlock()
	return mr.metrics[metricKey(name, labels)]
}

// ListMetrics возвращает все метрики
func (mr *MetricsRegistry) ListMetrics() []*Metric {
	mr.mutex.RLock()
	defer mr.mutex.RUnlock()
	result := make([]*Metric, 0, len(mr.metrics))
	for _, m := range mr.metrics {
		result = append(result, m)
	}
	return result
}

// ExportMetrics возвращает срез метрик для внешних систем мониторинга
func (mr *MetricsRegistry) ExportMetrics() []Metric {
	mr.mutex.RLock()
	defer mr.mutex.RUnlock()
	result := make([]Metric, 0, len(mr.metrics))
	for _, m := range mr.metrics {
		result = append(result, *m)
	}
	return result
}

// logMetric пишет метрику в лог-файл (если задан)
func (mr *MetricsRegistry) logMetric(metric *Metric) {
	if mr.logFile != nil {
		labels := ""
		for k, v := range metric.Labels {
			labels += k + "=" + v + " "
		}
		mr.logFile.WriteString(
			time.Now().Format(time.RFC3339) + " " +
				metric.Name + " [" + string(metric.Type) + "] " +
				labels + "value=" + formatFloat(metric.Value) + "\n",
		)
	}
}

// metricKey строит уникальный ключ для метрики по имени и лейблам
func metricKey(name string, labels map[string]string) string {
	key := name
	for k, v := range labels {
		key += "|" + k + "=" + v
	}
	return key
}

// formatFloat для красивого вывода
func formatFloat(f float64) string {
	if math.Abs(f-math.Round(f)) < 1e-9 {
		return fmt.Sprintf("%.0f", f)
	}
	return fmt.Sprintf("%.4f", f)
}

// Close закрывает лог-файл
func (mr *MetricsRegistry) Close() error {
	if mr.logFile != nil {
		return mr.logFile.Close()
	}
	return nil
}

// Пример использования:
/*
func main() {
	metrics, _ := NewMetricsRegistry("metrics.log")
	defer metrics.Close()
	metrics.SetGauge("tps", 120.5, map[string]string{"network": "mainnet"})
	metrics.IncCounter("blocks_created", 1, nil)
	metrics.ObserveHist("block_latency", 0.32, nil)
	m := metrics.GetMetric("tps", map[string]string{"network": "mainnet"})
	fmt.Println("TPS:", m.Value)
}
*/
