package ebiten

import (
	_ "embed"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed shader.kage
var shaderCode []byte

//go:embed mandelbox.kage
var mandelboxCode []byte

//go:embed mandelbulb.kage
var mandelbulbCode []byte

// initShader initializes the Kage shader
func (g *Game) initShader() {
	var err error
	fullShader := append([]byte(nil), shaderCode...)
	fullShader = append(fullShader, '\n')
	fullShader = append(fullShader, mandelboxCode...)
	fullShader = append(fullShader, '\n')
	fullShader = append(fullShader, mandelbulbCode...)

	g.shader, err = ebiten.NewShader(fullShader)
	if err != nil {
		log.Fatalf("Failed to compile shader: %v", err)
	}

	g.initialized = true
}
