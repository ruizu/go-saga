package saga

import (
	"context"
	"fmt"
	"sync"
)

// Step represents a single step in the saga
type Step struct {
	// Name is the name of the step
	Name string
	// Execute executes the step
	Execute func(ctx context.Context) error
	// Compensate compensates the step
	Compensate func(ctx context.Context) error
}

// Saga coordinates the distributed transaction
type Saga struct {
	steps      []Step
	mu         sync.Mutex
	executed   []Step
	onRollback func(context.Context, error)
}

// New creates a new saga instance
func New() *Saga {
	return &Saga{
		steps:    make([]Step, 0),
		executed: make([]Step, 0),
	}
}

// AddStep adds a new step to the saga
func (s *Saga) AddStep(step Step) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.steps = append(s.steps, step)
}

// OnRollback sets a callback function to be executed when rollback occurs
func (s *Saga) OnRollback(fn func(context.Context, error)) {
	s.onRollback = fn
}

// Execute runs all steps in the saga
func (s *Saga) Execute(ctx context.Context) error {
	for _, step := range s.steps {
		if err := s.executeStep(ctx, step); err != nil {
			rollbackErr := s.rollback(ctx)
			if rollbackErr != nil {
				return fmt.Errorf("execute failed: %v, rollback failed: %v", err, rollbackErr)
			}
			return fmt.Errorf("execute failed: %v, rollback successful", err)
		}
	}
	return nil
}

// executeStep executes a single step and records it
func (s *Saga) executeStep(ctx context.Context, step Step) error {
	if err := step.Execute(ctx); err != nil {
		return err
	}
	s.mu.Lock()
	s.executed = append(s.executed, step)
	s.mu.Unlock()
	return nil
}

// rollback compensates all executed steps in reverse order
func (s *Saga) rollback(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var rollbackErr error
	// Execute compensating transactions in reverse order
	for i := len(s.executed) - 1; i >= 0; i-- {
		step := s.executed[i]
		if err := step.Compensate(ctx); err != nil {
			rollbackErr = fmt.Errorf("compensation failed for step %s: %v", step.Name, err)
			break
		}
	}

	if s.onRollback != nil {
		s.onRollback(ctx, rollbackErr)
	}

	return rollbackErr
}
