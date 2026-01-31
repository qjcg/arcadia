package transition

import (
	"testing"
)

func TestZoomOutInTransition(t *testing.T) {
	allFractalTypes := []string{"mandelbrot", "julia", "burningship"}

	zoomOutIn := NewZoomOutInTransition(allFractalTypes)

	// Test initial state
	if zoomOutIn.Mode != TransitionZoomOutIn {
		t.Errorf("Expected mode %d, got %d", TransitionZoomOutIn, zoomOutIn.Mode)
	}

	// Test starting a transition
	zoomOutIn.Start("mandelbrot")

	if zoomOutIn.Target == "" {
		t.Error("Target should be set after Start")
	}

	if zoomOutIn.Target == "mandelbrot" {
		t.Error("Target should be different from current fractal")
	}

	// Test zoom out phase
	completed, zoomLevel, message := zoomOutIn.Update()

	if completed {
		t.Error("Transition should not be completed in zoom out phase")
	}

	// Zoom should be less than starting zoom (1.0)
	if zoomLevel >= 1.0 {
		t.Errorf("Expected zoom level less than 1.0 in zoom out phase, got %f", zoomLevel)
	}

	// Test zoom in phase - fast forward to middle of transition
	zoomOutIn.Progress = 0.6
	completed, zoomLevel, message = zoomOutIn.Update()

	if completed {
		t.Error("Transition should not be completed in zoom in phase")
	}

	// Test completion
	zoomOutIn.Progress = 0.95
	completed, zoomLevel, message = zoomOutIn.Update()

	if !completed {
		t.Error("Transition should be completed at end")
	}

	// Test message
	if message != "Transitioning to "+zoomOutIn.Target+" (Zoom Out/In)" {
		t.Errorf("Unexpected message: %s", message)
	}

	// Test GetMessage
	msg := zoomOutIn.GetMessage()
	expectedMsg := "Transitioning to " + zoomOutIn.Target + " (Zoom Out/In)"
	if msg != expectedMsg {
		t.Errorf("Expected message %s, got %s", expectedMsg, msg)
	}
}
