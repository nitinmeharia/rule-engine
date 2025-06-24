package execution

import (
	"testing"
	"time"
)

func TestCircuitBreaker_Execute(t *testing.T) {
	cb := &CircuitBreaker{
		maxFailures: 2,
		timeout:     100 * time.Millisecond,
		state:       CircuitBreakerClosed,
	}

	// Test successful execution
	err := cb.Execute(func() error {
		return nil
	})
	if err != nil {
		t.Errorf("Expected no error for successful execution, got: %v", err)
	}

	if cb.GetState() != CircuitBreakerClosed {
		t.Errorf("Expected circuit breaker to be closed, got: %v", cb.GetState())
	}

	// Test failed execution
	err = cb.Execute(func() error {
		return ErrServiceStopped
	})
	if err != ErrServiceStopped {
		t.Errorf("Expected service stopped error, got: %v", err)
	}

	if cb.GetFailures() != 1 {
		t.Errorf("Expected 1 failure, got: %d", cb.GetFailures())
	}

	// Test circuit breaker opening
	err = cb.Execute(func() error {
		return ErrServiceStopped
	})
	if err != ErrServiceStopped {
		t.Errorf("Expected service stopped error, got: %v", err)
	}

	if cb.GetState() != CircuitBreakerOpen {
		t.Errorf("Expected circuit breaker to be open, got: %v", cb.GetState())
	}

	// Test circuit breaker blocking execution when open
	err = cb.Execute(func() error {
		return nil
	})
	if err != ErrCircuitBreakerOpen {
		t.Errorf("Expected circuit breaker open error, got: %v", err)
	}
}

func TestCircuitBreaker_Recovery(t *testing.T) {
	cb := &CircuitBreaker{
		maxFailures: 1,
		timeout:     50 * time.Millisecond,
		state:       CircuitBreakerOpen,
		lastFailure: time.Now().Add(-100 * time.Millisecond), // Past timeout
	}

	// Test recovery after timeout
	err := cb.Execute(func() error {
		return nil
	})
	if err != nil {
		t.Errorf("Expected successful execution after timeout, got: %v", err)
	}

	if cb.GetState() != CircuitBreakerClosed {
		t.Errorf("Expected circuit breaker to be closed after recovery, got: %v", cb.GetState())
	}
}
