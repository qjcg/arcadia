package render

import (
	"strings"
	"testing"

	"github.com/qjcg/arcadia/exp/fractalis/internal/core/persistence"
	"github.com/qjcg/arcadia/exp/fractalis/internal/tui/color"
)

func stubCalculate(cr, ci float64, cfg persistence.Config) int {
	if cr*cr+ci*ci < 0.25 {
		return cfg.MaxIter
	}
	iter := int((cr + ci + 2) * float64(cfg.MaxIter) / 4)
	if iter < 0 {
		return 0
	}
	return iter
}

func testConfig(fractalType string) persistence.Config {
	return persistence.Config{
		Width:       20,
		Height:      10,
		MaxIter:     50,
		CenterX:     -0.5,
		CenterY:     0.0,
		Zoom:        1.0,
		ColorScheme: color.ColorGrayscale,
		FractalType: fractalType,
	}
}

func renderPlain(t *testing.T, cfg persistence.Config) string {
	t.Helper()
	r := NewRenderer(cfg, stubCalculate)
	r.SetColorMode(false)
	r.SetConfig(cfg)
	return r.RenderFractal()
}

func TestCrossfadeAtTheEndsShowsEachViewAlone(t *testing.T) {
	src := testConfig("mandelbrot")
	dst := testConfig("julia")
	dst.CenterX, dst.Zoom = 3.7, 4.2
	r := NewRenderer(src, stubCalculate)
	r.SetColorMode(false)

	if got, want := r.RenderCrossfade(src, dst, 0), renderPlain(t, src); got != want {
		t.Errorf("crossfade at mix 0 =\n%q\nwant the source view\n%q", got, want)
	}
	if got, want := r.RenderCrossfade(src, dst, 1), renderPlain(t, dst); got != want {
		t.Errorf("crossfade at mix 1 =\n%q\nwant the destination view\n%q", got, want)
	}
}

func TestCrossfadeDissolvesBetweenBothViews(t *testing.T) {
	src := testConfig("mandelbrot")
	dst := testConfig("julia")
	dst.CenterX, dst.Zoom = 3.7, 4.2
	srcOut := strings.Split(renderPlain(t, src), "\n")
	dstOut := strings.Split(renderPlain(t, dst), "\n")

	r := NewRenderer(src, stubCalculate)
	r.SetColorMode(false)
	got := strings.Split(r.RenderCrossfade(src, dst, 0.5), "\n")

	fromSrc, fromDst := 0, 0
	for row := range got {
		for col := range got[row] {
			switch got[row][col] {
			case srcOut[row][col]:
				fromSrc++
			case dstOut[row][col]:
				fromDst++
			default:
				t.Fatalf("cell (%d,%d) = %q, want a cell of the source or destination",
					row, col, got[row][col])
			}
		}
	}
	if fromSrc == 0 || fromDst == 0 {
		t.Fatalf("mid crossfade drew %d source and %d destination cells, want both",
			fromSrc, fromDst)
	}
}
