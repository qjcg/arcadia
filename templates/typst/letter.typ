#let letter(
  from-name: "Company Name",
  from-address: ("123 Main Street", "City, State 12345", "Country"),
  from-email: "hello@example.com",
  from-phone: "+1 (555) 123-4567",
  to-name: "Recipient Name",
  to-title: none,
  to-company: none,
  to-address: ("Street Address", "City, State ZIP", "Country"),
  date: datetime.today(),
  subject: "Subject Line",
  salutation: "Dear Sir or Madam,",
  closing: "Sincerely,",
  signatory: "Name of Signatory",
  signatory-title: "Title / Position",
  enclosures: none,
  body,
) = {
  let dark-blue = rgb("#2c3e50")
  let gold = rgb("#c0a060")
  let muted = rgb("#6b7b8d")
  let dark-gray = rgb("#555555")
  let beige = rgb("#f7f3eb")
  let tan = rgb("#d4c9a8")

  set document(title: subject)
  set page(
    paper: "us-letter",
    margin: (left: 2.5cm, right: 2.5cm, top: 2cm, bottom: 2cm),
    fill: white,
  )
  set text(font: "Libertinus Serif", size: 11pt)
  set par(justify: true, leading: 0.65em)

  // ── Thin decorative top rule ──
  line(length: 100%, stroke: 0.5pt + gold)
  v(-0.4em)

  // ── Sender block ──
  text(size: 16pt, weight: "bold", fill: dark-blue, [#from-name])
  v(0.15em)
  for line in from-address {
    text(size: 9.5pt, fill: gray, [#line])
    v(0.05em)
  }
  text(size: 9pt, fill: muted, [#from-email \ #from-phone])

  v(1.5em)

  // ── Date (right-aligned) ──
  align(right, text(size: 10.5pt, fill: dark-gray, [
    #date.display("[month repr:long] [day], [year]")
  ]))

  v(1.2em)

  // ── Recipient block ──
  text(size: 10.5pt, weight: "bold", [#to-name])
  v(0.05em)
  if to-title != none {
    text(size: 10pt, style: "italic", [#to-title])
    v(0.05em)
  }
  if to-company != none {
    text(size: 10pt, [#to-company])
    v(0.05em)
  }
  for line in to-address {
    text(size: 10pt, [#line])
    v(0.05em)
  }

  v(1.8em)

  // ── Subject line ──
  rect(
    width: 100%,
    fill: beige,
    inset: (x: 10pt, y: 6pt),
    stroke: 0.3pt + tan,
  )[
    text(size: 11pt, weight: "bold", fill: dark-blue, [Re: #subject])
  ]

  v(1.5em)

  // ── Salutation ──
  text(size: 11pt, [#salutation])

  v(1em)

  // ── Body ──
  body

  v(1.2em)

  // ── Closing ──
  text(size: 11pt, [#closing])

  v(2.2em)

  // ── Signatory ──
  line(length: 5cm, stroke: 0.5pt)
  v(0.15em)
  text(size: 11pt, weight: "bold", [#signatory])
  v(0.05em)
  text(size: 10pt, style: "italic", fill: gray, [#signatory-title])

  v(0.8em)

  // ── Enclosures ──
  if enclosures != none {
    line(length: 100%, stroke: 0.3pt + tan)
    v(0.4em)
    text(size: 9.5pt, weight: "bold", fill: dark-gray, [Enclosures:])
    for item in enclosures {
      v(0.05em)
      text(size: 9.5pt, fill: dark-gray, [— #item])
    }
  }

  // ── Thin decorative bottom rule ──
  v(1fr)
  line(length: 100%, stroke: 0.5pt + gold)
}
