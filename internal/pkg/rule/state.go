package rule

import (
	"golang.org/x/exp/constraints"
)

type OrderType int8

const (
	OrderTypeUnknown OrderType = iota
	OrderTypeNone
	OrderTypeAsc
	OrderTypeDesc
)

// SequenceState
type SequenceState[T constraints.Ordered] struct {
	last          *T
	expectedOrder OrderType
	Order         OrderType
}

func NewSequenceState[T constraints.Ordered](orderType OrderType) *SequenceState[T] {
	return &SequenceState[T]{
		expectedOrder: orderType,
		Order:         orderType,
	}
}

func (ss *SequenceState[T]) Add(value T) bool {
	if ss.last == nil {
		ss.last = &value
		return true
	}

	isOrdered := true
	switch ss.expectedOrder {
	case OrderTypeAsc:
		if value < *ss.last {
			isOrdered = false
			ss.Order = OrderTypeNone
		}
	case OrderTypeDesc:
		if value > *ss.last {
			isOrdered = false
			ss.Order = OrderTypeNone
		}
	}

	ss.last = &value

	return isOrdered
}

// CompleteState
type Number interface {
	constraints.Float | constraints.Integer
}

type CompleteState[T Number] struct {
	next        *T
	step        T
	AllComplete bool
}

func NewCompleteState[T Number](step T) *CompleteState[T] {
	return &CompleteState[T]{
		step:        step,
		AllComplete: true,
	}
}

func (ss *CompleteState[T]) Add(value T) bool {
	if ss.next == nil {
		ss.next = &value
		return true
	}

	isComplete := value == *ss.next
	if !isComplete {
		ss.AllComplete = false
	}

	*ss.next = value + ss.step

	return isComplete
}

// StateProvider
type StateProvider struct {
	// Sequence state
	SequenceEnabled bool
	sequenceOrder   OrderType
	sequenceInt64   *SequenceState[int64]
	sequenceUint64  *SequenceState[uint64]
	sequenceFloat64 *SequenceState[float64]
	sequenceString  *SequenceState[string]

	// Complete state
	CompleteEnabled bool
	completeStep    float64
	completeInt64   *CompleteState[int64]
	completeUint64  *CompleteState[uint64]
	completeFloat64 *CompleteState[float64]
}

func NewStateProvider() *StateProvider {
	return &StateProvider{}
}

func (sp *StateProvider) WithSequenceState(order OrderType) *StateProvider {
	sp.SequenceEnabled = true
	sp.sequenceOrder = order
	return sp
}

func (sp *StateProvider) WithCompleteState(step float64) *StateProvider {
	sp.CompleteEnabled = true
	sp.completeStep = step
	return sp
}

func (sp *StateProvider) SequenceAddInt64(v int64) bool {
	if sp.sequenceInt64 == nil {
		sp.sequenceInt64 = NewSequenceState[int64](sp.sequenceOrder)
	}
	return sp.sequenceInt64.Add(v)
}

func (sp *StateProvider) SequenceAddUint64(v uint64) bool {
	if sp.sequenceUint64 == nil {
		sp.sequenceUint64 = NewSequenceState[uint64](sp.sequenceOrder)
	}
	return sp.sequenceUint64.Add(v)
}

func (sp *StateProvider) SequenceAddFloat64(v float64) bool {
	if sp.sequenceFloat64 == nil {
		sp.sequenceFloat64 = NewSequenceState[float64](sp.sequenceOrder)
	}
	return sp.sequenceFloat64.Add(v)
}

func (sp *StateProvider) SequenceAddString(v string) bool {
	if sp.sequenceString == nil {
		sp.sequenceString = NewSequenceState[string](sp.sequenceOrder)
	}
	return sp.sequenceString.Add(v)
}

func (sp *StateProvider) CompleteAddInt64(v int64) bool {
	if sp.completeInt64 == nil {
		sp.completeInt64 = NewCompleteState[int64](int64(sp.completeStep))
	}
	return sp.completeInt64.Add(v)
}

func (sp *StateProvider) CompleteAddUint64(v uint64) bool {
	if sp.completeUint64 == nil {
		sp.completeUint64 = NewCompleteState[uint64](uint64(sp.completeStep))
	}
	return sp.completeUint64.Add(v)
}

func (sp *StateProvider) CompleteAddFloat64(v float64) bool {
	if sp.completeFloat64 == nil {
		sp.completeFloat64 = NewCompleteState[float64](sp.completeStep)
	}
	return sp.completeFloat64.Add(v)
}
