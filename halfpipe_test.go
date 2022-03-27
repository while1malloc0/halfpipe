package halfpipe_test

import (
	"context"
	"errors"
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

func TestBasic_error(t *testing.T) {
	pipeline := halfpipe.NewPipeline()
	f := func(ctx context.Context) (context.Context, error) {
		return ctx, errors.New("test error")
	}
	pipeline.MustAddStep("cause an error", &halfpipe.SerialPipelineStep{
		Action: f,
	})
	ctx := context.Background()
	_, err := pipeline.Run(ctx)
	assert.EqualError(t, err, "test error")
}

func TestCancellationIsRespected(t *testing.T) {
	pipeline := halfpipe.NewPipeline()
	pipeline.MustAddStep("cause an error", &noOpStep{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := pipeline.Run(ctx)
	assert.EqualError(t, err, "context canceled")
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
