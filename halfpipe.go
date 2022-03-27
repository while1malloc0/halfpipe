package halfpipe

import (
	"context"
	"fmt"
)

// Pipeline is the foundational type of the package. It is responsible for
// orchestrating the invocation of functions, retrying them if necessary,
// reporting errors, and ensuring that context cancellation is respected
type Pipeline struct {
	steps *stepMap
}

// NewPipeline returns a Pipeline with no steps included
func NewPipeline() *Pipeline {
	return &Pipeline{
		steps: newStepMap(),
	}
}

// MustAddStep enqueues a PipelineStep to be run by a Pipeline. This function
// panics if a duplicate step ID is given. For the non-panicing version, see
// AddStep.
func (p *Pipeline) MustAddStep(id string, step PipelineStep) {
	if err := p.AddStep(id, step); err != nil {
		panic(err)
	}
}

// AddStep enqueues a PipelineStep to be run by a Pipeline. It returns an error
// if a duplicate step ID is passed in.
func (p *Pipeline) AddStep(id string, step PipelineStep) error {
	if err := p.steps.Add(id, step); err != nil {
		return err
	}
	return nil
}

// Steps returns a list of step IDs in the order that the Pipeline will run them
func (p *Pipeline) Steps() []string {
	return p.steps.Keys()
}

// Run invokes each enqueued pipeline step
func (p *Pipeline) Run(ctx context.Context) (context.Context, error) {
	var err error
	for _, id := range p.Steps() {
		if ctx.Err() != nil {
			return ctx, ctx.Err()
		}
		ctx, err = p.steps.keyValues[id].Run(ctx)
		if err != nil {
			return ctx, err
		}
	}
	return ctx, nil
}

// PipelineStep represents any type that can be invoked as part of a Pipeline
type PipelineStep interface {
	Run(context.Context) (context.Context, error)
}

// RunFunc is the basic function type that gets invoked as part of a Pipeline
type RunFunc func(context.Context) (context.Context, error)

// SerialPipelineStep is a PipelineStep that invokes a single function
type SerialPipelineStep struct {
	Action RunFunc
}

// Run implements the PipelineStep interface by invoking the Action function
func (step *SerialPipelineStep) Run(ctx context.Context) (context.Context, error) {
	return step.Action(ctx)
}

// stepMap is an domain-specific ordered map of (step id, PipelineStep) pairs.
// Instead of overriding duplicate keys, a stepMap will return an error if you
// attempt to add one.
// Note: the error on duplication guarantee is only enforceable as long as this
// struct's fields are not accessed directly.
type stepMap struct {
	keys      []string
	keyValues map[string]PipelineStep
}

func newStepMap() *stepMap {
	return &stepMap{
		keys:      []string{},
		keyValues: make(map[string]PipelineStep),
	}
}

// Add stores a new key and value in the map. If a duplicate key is given, an
// error is returned
func (sm *stepMap) Add(key string, value PipelineStep) error {
	if contains(sm.keys, key) {
		return fmt.Errorf("duplicate id for step: %s", key)
	}
	sm.keys = append(sm.keys, key)
	sm.keyValues[key] = value
	return nil
}

// Keys returns the list of map keys in the order that they were inserted
func (sm *stepMap) Keys() []string {
	return sm.keys
}

func contains(haystack []string, needle string) bool {
	for i := range haystack {
		if haystack[i] == needle {
			return true
		}
	}
	return false
}
