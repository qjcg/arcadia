package transition

import (
	"testing"
)

func TestBreakthroughTransition(t *testing.T) {
	allFractalTypes := []string{"mandelbrot", "julia", "burningship"}

	breakthrough := NewBreakthroughTransition(allFractalTypes)

	// Test initial state
	if breakthrough.Mode != TransitionBreakthrough {
		t.Errorf("Expected mode %d, got %d", TransitionBreakthrough, breakthrough.Mode)
	}

	// Test starting a transition
	breakthrough.Start("mandelbrot")

	if breakthrough.Target == "" {
		t.Error("Target should be set after Start")
	}

	if breakthrough.Target == "mandelbrot" {
		t.Error("Target should be different from current fractal")
	}

	// Test phase 1 - zoom out
	breakthrough.Progress = 0.1
	completed, centerX, centerY, zoomLevel, message := breakthrough.Update()

	if completed {
		t.Error("Transition should not be completed in phase 1")
	}

	// Zoom should be decreasing
	if zoomLevel >= 1.0 {
		t.Errorf("Expected zoom level less than 1.0 in phase 1, got %f", zoomLevel)
	}

	// Test phase 2 - position shift and zoom in
	breakthrough.Progress = 0.4
	completed, centerX, centerY, zoomLevel, message = breakthrough.Update()

	if completed {
		t.Error("Transition should not be completed in phase 2")
	}

	// Position should be changing dramatically
	if centerX == -0.5 && centerY == 0.0 {
		t.Error("Position should be changing in phase 2")
	}

	// Test phase 3 - final stabilization
	breakthrough.Progress = 0.7
	completed, centerX, centerY, zoomLevel, message = breakthrough.Update()

	if completed {
		t.Error("Transition should not be completed in phase 3")
	}

	// Test completion
	breakthrough.Progress = 0.95
	completed, centerX, centerY, zoomLevel, message = breakthrough.Update()

	if !completed {
		t.Error("Transition should be completed at end")
	}

	// Test message
	if message != "Transitioning to "+breakthrough.Target+" (Breakthrough)" {
		t.Errorf("Unexpected message: %s", message)
	}

	// Test GetMessage
	msg := breakthrough.GetMessage()
	expectedMsg := "Transitioning to " + breakthrough.Target + " (Breakthrough)"
	if msg != expectedMsg {
		t.Errorf("Expected message %s, got %s", expectedMsg, msg)
	}
}
