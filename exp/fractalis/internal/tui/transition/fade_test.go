package transition

import (
	"testing"
)

func TestFadeTransition(t *testing.T) {
	allFractalTypes := []string{"mandelbrot", "julia", "burningship"}

	fade := NewFadeTransition(allFractalTypes)

	// Test initial state
	if fade.Mode != TransitionFade {
		t.Errorf("Expected mode %d, got %d", TransitionFade, fade.Mode)
	}

	if fade.Progress != 0.0 {
		t.Errorf("Expected initial progress 0.0, got %f", fade.Progress)
	}

	// Test starting a transition
	fade.Start("mandelbrot")

	if fade.Target == "" {
		t.Error("Target should be set after Start")
	}

	if fade.Target == "mandelbrot" {
		t.Error("Target should be different from current fractal")
	}

	if fade.Progress != 0.01 {
		t.Errorf("Expected progress 0.01 after Start, got %f", fade.Progress)
	}

	// Test updating the transition
	completed, message := fade.Update()

	if completed {
		t.Error("Transition should not be completed after one update")
	}

	if fade.Progress < 0.05 || fade.Progress > 0.07 {
		t.Errorf("Expected progress around 0.06 after one update, got %f", fade.Progress)
	}

	// Test message
	if message != "" {
		t.Errorf("Expected empty message during transition, got %s", message)
	}

	// Test completion
	for i := 0; i < 20; i++ {
		completed, _ = fade.Update()
	}

	if !completed {
		t.Error("Transition should be completed after many updates")
	}

	// Test GetMessage
	msg := fade.GetMessage()
	expectedMsg := "Transitioning to " + fade.Target + " (Fade)"
	if msg != expectedMsg {
		t.Errorf("Expected message %s, got %s", expectedMsg, msg)
	}
}
