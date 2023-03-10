package global

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/neblic/platform/sampler/defs"
	"google.golang.org/protobuf/proto"
)

/*
 * Based on the go.opentelemetry.io/otel/internal/global package
 */

type samplerProviderPlaceholder struct {
	sync.Mutex
	samplers map[string]*samplerPlaceholder

	delegate defs.Provider
}

var _ defs.Provider = &samplerProviderPlaceholder{}

func newProveProviderPlaceholder() *samplerProviderPlaceholder {
	return &samplerProviderPlaceholder{}
}

func (p *samplerProviderPlaceholder) setDelegate(pp defs.Provider) error {
	p.Lock()
	defer p.Unlock()

	p.delegate = pp

	if len(p.samplers) == 0 {
		return nil
	}

	var aggErr error
	for _, pph := range p.samplers {
		if err := pph.setDelegate(pp); err != nil {
			if aggErr != nil {
				aggErr = fmt.Errorf("%s; %w", aggErr, err)
			} else {
				aggErr = err
			}
		}
	}

	p.samplers = nil

	return aggErr
}

func (p *samplerProviderPlaceholder) Sampler(name string, schema defs.Schema) (defs.Sampler, error) {
	p.Lock()
	defer p.Unlock()

	if p.delegate != nil {
		return p.delegate.Sampler(name, schema)
	}

	if p.samplers == nil {
		p.samplers = make(map[string]*samplerPlaceholder)
	}

	if val, ok := p.samplers[name]; ok {
		return val, nil
	}

	pw := &samplerPlaceholder{name: name, schema: schema}
	p.samplers[name] = pw

	return pw, nil
}

var _ defs.Sampler = &samplerPlaceholder{}

type samplerPlaceholder struct {
	name   string
	schema defs.Schema

	delegate atomic.Value
}

func (pw *samplerPlaceholder) setDelegate(pp defs.Provider) error {
	sampler, err := pp.Sampler(pw.name, pw.schema)
	if err != nil {
		return fmt.Errorf("error creating sampler: %w", err)
	}
	pw.delegate.Store(sampler)

	return nil
}

func (pw *samplerPlaceholder) SampleJSON(ctx context.Context, jsonSample string) (bool, error) {
	delegate := pw.delegate.Load()
	if delegate != nil {
		return delegate.(defs.Sampler).SampleJSON(ctx, jsonSample)
	}

	return false, nil
}

func (pw *samplerPlaceholder) SampleNative(ctx context.Context, nativeSample any) (bool, error) {
	delegate := pw.delegate.Load()
	if delegate != nil {
		return delegate.(defs.Sampler).SampleNative(ctx, nativeSample)
	}

	return false, nil
}

func (pw *samplerPlaceholder) SampleProto(ctx context.Context, protoSample proto.Message) (bool, error) {
	delegate := pw.delegate.Load()
	if delegate != nil {
		return delegate.(defs.Sampler).SampleProto(ctx, protoSample)
	}

	return false, nil
}

func (pw *samplerPlaceholder) Close() error {
	delegate := pw.delegate.Load()
	if delegate != nil {
		return delegate.(defs.Sampler).Close()
	}

	return nil
}
