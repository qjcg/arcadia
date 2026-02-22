package main

// DefaultTemplate is the default diploma template.
var DefaultTemplate = &Template{
	Image: "logo.svg",
	Overlay: map[string][2]float64{
		"Recipient":  {300, 200},
		"Course":     {300, 240},
		"Period":     {300, 260},
		"Instructor": {300, 280},
		"Image":      {100, 10},
	},
}
