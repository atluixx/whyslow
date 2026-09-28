package analyzers

import "github.com/atluixx/whyslow/internal/models"

// History is bounded in-memory state used to recognize sustained pressure and
// retain recent events. It is intentionally not persisted.
type History struct {
	capacity int
	samples  []models.Sample
	events   []models.Event
	active   map[string]bool
}

func NewHistory(capacity int) *History {
	if capacity < 1 {
		capacity = 1
	}
	return &History{capacity: capacity, active: make(map[string]bool)}
}

func (h *History) AddSample(sample models.Sample) {
	h.samples = append(h.samples, sample)
	if len(h.samples) > h.capacity {
		h.samples = h.samples[len(h.samples)-h.capacity:]
	}
}

func (h *History) Samples() []models.Sample { return append([]models.Sample(nil), h.samples...) }
func (h *History) Events() []models.Event   { return append([]models.Event(nil), h.events...) }

func (h *History) Sustained(count int, predicate func(models.Sample) bool) bool {
	if count < 1 || len(h.samples) < count {
		return false
	}
	for _, sample := range h.samples[len(h.samples)-count:] {
		if !predicate(sample) {
			return false
		}
	}
	return true
}

// SetCondition emits only when a condition first becomes true. Clearing the
// condition permits a new event if it returns later.
func (h *History) SetCondition(key string, active bool, event models.Event) bool {
	if !active {
		delete(h.active, key)
		return false
	}
	if h.active[key] {
		return false
	}
	h.active[key] = true
	h.events = append(h.events, event)
	if len(h.events) > h.capacity {
		h.events = h.events[len(h.events)-h.capacity:]
	}
	return true
}
