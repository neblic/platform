package function

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/cel-go/common/types/ref"
)

var ErrMaxKeys = errors.New("max keys reached")

type StatefulFunction interface {
	ref.Type
	ref.Val
	Call(value any) bool
}

type state struct {
	state        any
	lastAccessed time.Time
}

type StatefulFunctionProvider struct {
	StateName               string
	mutex                   sync.RWMutex
	ctx                     context.Context
	ctxCancel               context.CancelFunc
	maxKeys                 int32
	statefulFunctionBuilder func(state any) StatefulFunction
	stateBuilder            func() any
	globalStatefulFunction  StatefulFunction
	keyedStates             map[string]state
}

func NewStatefulFunctionProvider(stateName string, statefulFunctionBuilder func(state any) StatefulFunction, stateBuilder func() any) *StatefulFunctionProvider {
	ctx, cancel := context.WithCancel(context.Background())
	return &StatefulFunctionProvider{
		StateName:               stateName,
		mutex:                   sync.RWMutex{},
		ctx:                     ctx,
		ctxCancel:               cancel,
		statefulFunctionBuilder: statefulFunctionBuilder,
		stateBuilder:            stateBuilder,
		keyedStates:             make(map[string]state),
	}
}

func (sfp *StatefulFunctionProvider) WithManagedKeyedState(ttl time.Duration, maxKeys int32) {
	sfp.maxKeys = maxKeys

	ticker := time.NewTicker(time.Minute)
	go func() {
		for {
			select {
			case <-sfp.ctx.Done():
				return
			case <-ticker.C:
				for key, state := range sfp.keyedStates {
					if time.Since(state.lastAccessed) > ttl {
						sfp.mutex.Lock()
						delete(sfp.keyedStates, key)
						sfp.mutex.Unlock()
					}
				}
			}
		}
	}()
}

func (sfp *StatefulFunctionProvider) GlobalStatefulFunction() StatefulFunction {
	if sfp.globalStatefulFunction == nil {
		sfp.globalStatefulFunction = sfp.statefulFunctionBuilder(sfp.stateBuilder())
	}
	return sfp.globalStatefulFunction
}

func (sfp *StatefulFunctionProvider) KeyedStatefulFunction(key string) (StatefulFunction, error) {
	sfp.mutex.RLock()
	s, ok := sfp.keyedStates[key]
	sfp.mutex.RUnlock()
	if !ok {

		// Check if the maximum number of keys has been reached
		if sfp.maxKeys > 0 && int32(len(sfp.keyedStates)) >= sfp.maxKeys {
			return nil, ErrMaxKeys
		}

		s = state{
			state: sfp.stateBuilder(),
		}
		sfp.mutex.Lock()
		sfp.keyedStates[key] = s
		sfp.mutex.Unlock()
	}
	s.lastAccessed = time.Now()

	return sfp.statefulFunctionBuilder(s.state), nil
}

func (sfp *StatefulFunctionProvider) Close() {
	sfp.ctxCancel()
}
