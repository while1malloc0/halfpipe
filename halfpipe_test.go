package halfpipe_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/while1malloc0/halfpipe"
)

type testStep func(context.Context) (context.Context, error)

func (s testStep) Run(ctx context.Context) (context.Context, error) {
	return s(ctx)
}

type noOpStep struct{}

func (noop *noOpStep) Run(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func noop() halfpipe.PipelineStep {
	return &noOpStep{}
}

func TestBasic(t *testing.T) {
	subject := halfpipe.NewPipeline()
	tf1 := func(ctx context.Context) (context.Context, error) {
		return context.WithValue(ctx, "tfcalled1", true), nil
	}
	tf2 := func(ctx context.Context) (context.Context, error) {
		return context.WithValue(ctx, "tfcalled2", true), nil
	}
	subject.MustAddStep("tf1", &halfpipe.SerialPipelineStep{
		Action: tf1,
	})
	subject.MustAddStep("tf2", &halfpipe.SerialPipelineStep{
		Action: tf2,
	})
	ctx := context.Background()
	got, err := subject.Run(ctx)
	assert.Nil(t, err)
	assert.True(t, got.Value("tfcalled1").(bool))
	assert.True(t, got.Value("tfcalled2").(bool))
}

func TestBasic_error(t *testing.T) {
	subject := halfpipe.NewPipeline()
	f := func(ctx context.Context) (context.Context, error) {
		return ctx, errors.New("test error")
	}
	subject.MustAddStep("cause an error", &halfpipe.SerialPipelineStep{
		Action: f,
	})
	ctx := context.Background()
	_, err := subject.Run(ctx)
	assert.EqualError(t, err, "test error")
}

func TestCancellationIsRespected(t *testing.T) {
	subject := halfpipe.NewPipeline()
	subject.MustAddStep("cause an error", &noOpStep{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := subject.Run(ctx)
	assert.EqualError(t, err, "context canceled")
}

func TestContextIsPassedBetweenSteps(t *testing.T) {
	subject := halfpipe.NewPipeline()
	subject.MustAddStep("write to context", testStep(func(ctx context.Context) (context.Context, error) {
		written := context.WithValue(ctx, "written-to", true)
		return written, nil
	}))
	subject.MustAddStep("read from context", testStep(func(ctx context.Context) (context.Context, error) {
		if !ctx.Value("written-to").(bool) {
			return nil, errors.New("didn't find expected context value for key written-to")
		}
		return ctx, nil
	}))
	ctx := context.Background()
	got, err := subject.Run(ctx)
	assert.Nil(t, err)
	assert.True(t, got.Value("written-to").(bool))
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
