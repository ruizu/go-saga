package saga

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSaga(t *testing.T) {
	ctx := context.Background()
	s := New()
	s.OnRollback(func(ctx context.Context, err error) {
	})
	s.AddStep(Step{
		Name: "step1",
		Execute: func(ctx context.Context) error {
			return nil
		},
		Compensate: func(ctx context.Context) error {
			return nil
		},
	})

	err := s.Execute(ctx)
	assert.NoError(t, err)
}

func TestSagaRollback(t *testing.T) {
	ctx := context.Background()
	s := New()
	s.OnRollback(func(ctx context.Context, err error) {
	})
	s.AddStep(Step{
		Name: "step1",
		Execute: func(ctx context.Context) error {
			return nil
		},
		Compensate: func(ctx context.Context) error {
			return nil
		},
	})
	s.AddStep(Step{
		Name: "step2",
		Execute: func(ctx context.Context) error {
			return errors.New("error step 2")
		},
		Compensate: func(ctx context.Context) error {
			return nil
		},
	})
	s.AddStep(Step{
		Name: "step3",
		Execute: func(ctx context.Context) error {
			return nil
		},
		Compensate: func(ctx context.Context) error {
			return nil
		},
	})

	err := s.Execute(ctx)
	if assert.Error(t, err) {
		assert.Equal(t, "execute failed: error step 2, rollback successful", err.Error())
	}
}

func TestSagaErrorCompensate(t *testing.T) {
	ctx := context.Background()
	s := New()
	s.OnRollback(func(ctx context.Context, err error) {
	})
	s.AddStep(Step{
		Name: "step1",
		Execute: func(ctx context.Context) error {
			return nil
		},
		Compensate: func(ctx context.Context) error {
			return errors.New("error compensate step 1")
		},
	})
	s.AddStep(Step{
		Name: "step2",
		Execute: func(ctx context.Context) error {
			return errors.New("error step 2")
		},
		Compensate: func(ctx context.Context) error {
			return nil
		},
	})

	err := s.Execute(ctx)
	if assert.Error(t, err) {
		assert.Equal(t, "execute failed: error step 2, rollback failed: compensation failed for step step1: error compensate step 1", err.Error())
	}
}
