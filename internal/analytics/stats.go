package analytics

import (
	"math"
	"sync"
)

// Statistics предоставляет методы для статистических вычислений
type Statistics struct {
	mu sync.RWMutex
}

// NewStatistics создает новый экземпляр Statistics
func NewStatistics() *Statistics {
	return &Statistics{}
}

// CalculateMean вычисляет среднее значение
func (s *Statistics) CalculateMean(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}

	var sum float64
	for _, v := range data {
		sum += v
	}
	return sum / float64(len(data))
}

// CalculateStdDev вычисляет стандартное отклонение
func (s *Statistics) CalculateStdDev(data []float64, mean float64) float64 {
	if len(data) < 2 {
		return 0
	}

	var sum float64
	for _, v := range data {
		diff := v - mean
		sum += diff * diff
	}

	return math.Sqrt(sum / float64(len(data)-1))
}

// CalculateZScore вычисляет Z-score для значения
func (s *Statistics) CalculateZScore(value, mean, stdDev float64) float64 {
	if stdDev == 0 {
		return 0
	}
	return (value - mean) / stdDev
}

// RollingAverage вычисляет скользящее среднее
func (s *Statistics) RollingAverage(data []float64, windowSize int) []float64 {
	if len(data) == 0 || windowSize <= 0 {
		return []float64{}
	}

	result := make([]float64, len(data))

	for i := range data {
		start := max(0, i-windowSize+1)
		window := data[start : i+1]

		var sum float64
		for _, v := range window {
			sum += v
		}
		result[i] = sum / float64(len(window))
	}

	return result
}

// DetectAnomalies обнаруживает аномалии с помощью Z-score
func (s *Statistics) DetectAnomalies(data []float64, threshold float64) []bool {
	if len(data) == 0 {
		return []bool{}
	}

	mean := s.CalculateMean(data)
	stdDev := s.CalculateStdDev(data, mean)

	anomalies := make([]bool, len(data))

	for i, value := range data {
		zScore := s.CalculateZScore(value, mean, stdDev)
		anomalies[i] = math.Abs(zScore) > threshold
	}

	return anomalies
}

// SimpleExponentialSmoothing применяет простое экспоненциальное сглаживание
func (s *Statistics) SimpleExponentialSmoothing(data []float64, alpha float64) []float64 {
	if len(data) == 0 {
		return []float64{}
	}

	smoothed := make([]float64, len(data))
	smoothed[0] = data[0]

	for i := 1; i < len(data); i++ {
		smoothed[i] = alpha*data[i] + (1-alpha)*smoothed[i-1]
	}

	return smoothed
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
