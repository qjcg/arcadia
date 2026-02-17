#let spec(
  title: none,
  author: none,
  body,
) = {
  set document(title: title) if title != none
  set page(
    paper: "us-letter",
    margin: 2.5cm,
    numbering: "1",
  )
  set text(size: 11pt)
  set par(justify: true)
  set heading(numbering: "1.")

  // Title page
  page(numbering: none, footer: none)[
    #block(
      width: 100%,
      height: 100%,
      stroke: 2pt + blue,
      inset: 1cm,
    )[
      #set align(center + horizon)
      #if title != none {
        text(24pt, weight: "bold", title)
        v(1em)
      }
      #if author != none {
        text(14pt, weight: "semibold", author)
        v(2em)
      }
      #text(10pt, gray, [Generated with Typst])
    ]
  ]

  // TOC
  page(numbering: none, footer: none)[
    #align(center, text(16pt, weight: "bold", [Table of Contents]))
    #v(1em)
    #outline(title: none, depth: 3)
  ]

  body

  // Final page
  page(numbering: none, footer: none)[
    #set align(center + horizon)
    #text(20pt, weight: "bold", fill: gray, [End of Specification Document])
  ]
}
