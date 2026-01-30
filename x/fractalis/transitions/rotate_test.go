package transitions

import (
	"testing"
)

func TestRotateTransition(t *testing.T) {
	allFractalTypes := []string{"mandelbrot", "julia", "burningship"}

	rotate := NewRotateTransition(allFractalTypes)

	// Test initial state
	if rotate.Mode != TransitionRotate {
		t.Errorf("Expected mode %d, got %d", TransitionRotate, rotate.Mode)
	}

	// Test starting a transition
	rotate.Start("mandelbrot")

	if rotate.Target == "" {
		t.Error("Target should be set after Start")
	}

	if rotate.Target == "mandelbrot" {
		t.Error("Target should be different from current fractal")
	}

	// Test rotation updates
	completed, centerX, centerY, message := rotate.Update()

	if completed {
		t.Error("Transition should not be completed immediately")
	}

	// Center should be different from initial (-0.5, 0.0)
	if centerX == -0.5 && centerY == 0.0 {
		t.Error("Center should change during rotation")
	}

	// Test completion
	rotate.Progress = 0.95
	completed, centerX, centerY, message = rotate.Update()

	if !completed {
		t.Error("Transition should be completed at end")
	}

	// Final position should be back to default
	if centerX != -0.5 || centerY != 0.0 {
		t.Errorf("Expected final position (-0.5, 0.0), got (%f, %f)", centerX, centerY)
	}

	// Test message
	if message != "Transitioning to "+rotate.Target+" (Rotate)" {
		t.Errorf("Unexpected message: %s", message)
	}

	// Test GetMessage
	msg := rotate.GetMessage()
	expectedMsg := "Transitioning to " + rotate.Target + " (Rotate)"
	if msg != expectedMsg {
		t.Errorf("Expected message %s, got %s", expectedMsg, msg)
	}
}
