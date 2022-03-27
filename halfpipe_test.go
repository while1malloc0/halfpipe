package halfpipe_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/while1malloc0/halfpipe"
)

type noOpStep struct{}

func (noop *noOpStep) Run(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func noop() halfpipe.PipelineStep {
	return &noOpStep{}
}

func TestBasic(t *testing.T) {
	pipeline := halfpipe.NewPipeline()
	tf1 := func(ctx context.Context) (context.Context, error) {
		return context.WithValue(ctx, "tfcalled1", true), nil
	}
	tf2 := func(ctx context.Context) (context.Context, error) {
		return context.WithValue(ctx, "tfcalled2", true), nil
	}
	pipeline.MustAddStep("tf1", &halfpipe.SerialPipelineStep{
		Action: tf1,
	})
	pipeline.MustAddStep("tf2", &halfpipe.SerialPipelineStep{
		Action: tf2,
	})
	ctx := context.Background()
	resultCtx, err := pipeline.Run(ctx)
	assert.Nil(t, err)
	assert.True(t, resultCtx.Value("tfcalled1").(bool))
	assert.True(t, resultCtx.Value("tfcalled2").(bool))
}

func TestAddStep(t *testing.T) {
	subject := halfpipe.NewPipeline()
	assert.NoError(t, subject.AddStep("step", noop()))
	assert.Contains(t, subject.Steps(), "step")
}

func TestAddStep_duplicate(t *testing.T) {
	subject := halfpipe.NewPipeline()
	assert.NoError(t, subject.AddStep("step", noop()))
	assert.Error(t, subject.AddStep("step", noop()))
}
