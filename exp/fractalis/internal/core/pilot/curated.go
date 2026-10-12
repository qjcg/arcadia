package pilot

// CuratedViewpoints returns the journey the Vantage pilot tours:
// curated compositions chosen to reliably show the most striking
// features each fractal type offers.
func CuratedViewpoints() []Viewpoint {
	return []Viewpoint{
		{
			Name: "Spiral Region",
			View: View{
				FractalType: "mandelbrot",
				CenterX:     -0.7436,
				CenterY:     0.1314,
				Zoom:        400,
				MaxIter:     240,
			},
		},
		{
			Name: "Elephant Valley",
			View: View{
				FractalType: "mandelbrot",
				CenterX:     -0.16,
				CenterY:     1.0405,
				Zoom:        150,
				MaxIter:     200,
			},
		},
		{
			Name: "Seahorse Valley",
			View: View{
				FractalType: "mandelbrot",
				CenterX:     -0.7269,
				CenterY:     0.1889,
				Zoom:        300,
				MaxIter:     220,
			},
		},
		{
			Name: "Double Spiral",
			View: View{
				FractalType: "mandelbrot",
				CenterX:     -0.7453,
				CenterY:     0.1127,
				Zoom:        1500,
				MaxIter:     300,
			},
		},
		{
			Name: "East Mini-Mandelbrot",
			View: View{
				FractalType: "mandelbrot",
				CenterX:     0.2805,
				CenterY:     0.0089,
				Zoom:        350,
				MaxIter:     220,
			},
		},
		{
			Name: "Douady Rabbit",
			View: View{
				FractalType: "julia",
				CenterX:     0,
				CenterY:     0,
				Zoom:        2,
				MaxIter:     150,
				JuliaCr:     0.285,
				JuliaCi:     0.01,
			},
		},
		{
			Name: "San Marco Dragon",
			View: View{
				FractalType: "julia",
				CenterX:     0,
				CenterY:     0,
				Zoom:        2,
				MaxIter:     150,
				JuliaCr:     -0.8,
				JuliaCi:     0.156,
			},
		},
		{
			Name: "Siegel Disk",
			View: View{
				FractalType: "julia",
				CenterX:     0,
				CenterY:     0,
				Zoom:        3,
				MaxIter:     150,
				JuliaCr:     -0.835,
				JuliaCi:     -0.2321,
			},
		},
		{
			Name: "Burning Ship",
			View: View{
				FractalType: "burningship",
				CenterX:     -0.5,
				CenterY:     -0.6,
				Zoom:        25,
				MaxIter:     150,
			},
		},
		{
			Name: "Tricorn",
			View: View{
				FractalType: "tricorn",
				CenterX:     -0.5,
				CenterY:     0,
				Zoom:        5,
				MaxIter:     120,
			},
		},
	}
}
