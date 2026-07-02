#let diploma(
  student: "Student Name",
  course: "Course Name",
  instructor: "Instructor Name",
  dates: "February 2026",
  recognition_title: "OFFICIAL RECOGNITION",
  diploma_title: "Diploma of Achievement",
  authority_text: "Presented by the authority of the lead instructor and the administration",
  signatory_title: "Lead Instructor",
  seal_text: "VERITAS & VIRTUS",
) = {
  set page(
    paper: "us-letter",
    flipped: true,
    margin: 1cm,
    background: place(center + horizon, rotate(25deg, text(
      120pt,
      fill: rgb("#f5f5f5"),
    )[CERTIFICATE])),
  )

  set text(font: "Libertinus Serif", size: 12pt)

  // Combined borders
  rect(
    width: 100%,
    height: 100%,
    stroke: 4pt + rgb("#8b4513"),
    inset: 20pt,
  )[
    #rect(
      width: 100%,
      height: 100%,
      stroke: 1.5pt + rgb("#d4af37"),
      inset: 30pt,
    )[
      #set align(center)
      #v(0.2fr)

      #text(size: 14pt, tracking: 3pt, weight: "bold", fill: rgb("#8b4513"))[
        #recognition_title
      ]

      #v(0.3cm)

      #text(size: 42pt, weight: "bold", fill: rgb("#2c3e50"))[
        #diploma_title
      ]

      #v(0.5cm)

      #text(size: 18pt, style: "italic")[
        This certifies that
      ]

      #v(0.2cm)

      #text(size: 34pt, weight: "bold", fill: black)[
        #student
      ]

      #v(0.2cm)

      #text(size: 18pt, style: "italic")[
        has demonstrated proficiency and successfully completed
      ]

      #v(0.2cm)

      #text(size: 24pt, weight: "semibold", fill: rgb("#2c3e50"))[
        #course
      ]

      #v(0.5cm)

      #text(size: 12pt)[
        #authority_text
      ]

      // Push signature bar to the very bottom
      #v(1fr)

      #grid(
        columns: (1.2fr, 1fr, 1.2fr),
        align: bottom,
        gutter: 15pt,
        // Instructor/Signatory on the left
        align(left + bottom)[
          #line(length: 100%, stroke: 0.5pt)
          #text(size: 13pt, weight: "semibold")[#instructor] \
          #text(size: 10pt, style: "italic")[#signatory_title]
        ],
        // Date in the middle
        [
          #line(length: 100%, stroke: 0.5pt)
          #text(size: 13pt, weight: "semibold")[#dates] \
          #text(size: 10pt, style: "italic")[Awarded Date]
        ],
        // Seal on the right
        align(right + bottom)[
          #let r = 38pt
          #let n = 48
          #box(width: r * 2, height: r * 2)[
            // Base polygon: centered by offsetting coordinates by r
            #polygon(
              fill: rgb("#d4af37"),
              stroke: 0.6pt + rgb("#b8860b"),
              ..range(n).map(i => {
                let angle = i * 360deg / n
                let r-i = if calc.even(i) { r } else { r * 0.92 }
                (r + calc.cos(angle) * r-i, r + calc.sin(angle) * r-i)
              }),
            )

            // Place text block at the center of the r*2 box
            #place(center + horizon)[
              #set text(
                size: 7.2pt,
                fill: white.darken(5%),
                weight: "bold",
                tracking: 0.4pt,
              )
              #set par(leading: 0.35em)
              #box(width: r * 1.5)[
                #align(center, seal_text)
              ]
            ]
          ]
        ],
      )

      #v(0.1fr)
    ]
  ]
}
