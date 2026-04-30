package graph

import (
	"sync"
	"sync/atomic"
)

// StatsCollector maintains runtime statistics for query optimization.
// It tracks approximate counts of nodes per label to help the query
// optimizer choose the most selective index.
type StatsCollector struct {
	mu       sync.RWMutex
	counts   map[string]int64
	enabled  int32 // atomic bool
}

// NewStatsCollector creates a new StatsCollector.
func NewStatsCollector() *StatsCollector {
	return &StatsCollector{
		counts:  make(map[string]int64),
		enabled: 1,
	}
}

// RecordLabel records that a node with the given label was added.
func (s *StatsCollector) RecordLabel(label string) {
	if s == nil || atomic.LoadInt32(&s.enabled) == 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counts[label]++
}

// RemoveLabel records that a node with the given label was removed.
func (s *StatsCollector) RemoveLabel(label string) {
	if s == nil || atomic.LoadInt32(&s.enabled) == 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.counts[label] > 0 {
		s.counts[label]--
	}
}

// EstimateCount returns the estimated number of nodes with the given label.
func (s *StatsCollector) EstimateCount(label string) int64 {
	if s == nil {
		return 0
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.counts[label]
}

// SelectBestLabel chooses the label with the smallest estimated count
// from the given candidates. This helps the query optimizer pick the
// most selective index.
func (s *StatsCollector) SelectBestLabel(labels []string) string {
	if s == nil || len(labels) == 0 {
		return ""
	}
	if len(labels) == 1 {
		return labels[0]
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	bestLabel := labels[0]
	bestCount := s.counts[bestLabel]
	if bestCount == 0 {
		bestCount = 1 << 60 // treat unknown as very large
	}

	for i := 1; i < len(labels); i++ {
		count := s.counts[labels[i]]
		if count == 0 {
			count = 1 << 60
		}
		if count < bestCount {
			bestCount = count
			bestLabel = labels[i]
		}
	}
	return bestLabel
}

// Disable disables statistics collection.
func (s *StatsCollector) Disable() {
	atomic.StoreInt32(&s.enabled, 0)
}

// Enable enables statistics collection.
func (s *StatsCollector) Enable() {
	atomic.StoreInt32(&s.enabled, 1)
}
