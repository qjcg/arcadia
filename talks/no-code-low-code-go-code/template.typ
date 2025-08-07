#let theme = (
    paper: "presentation-16-9",

	baseColorBackground: black,
	baseColorForeground: white,
	baseFontFamily: "Jost",
	baseSize: 20pt,
	baseSizeHeading1: 80pt,
	baseSizeHeading2: 40pt,
    baseSizeFooter: 10pt,

    titleColorForeground: lime,
    titleSize: 80pt,
	titleSizeSubtitle: 20pt,
	titleSizeMetadata: 15pt,

	rawFontFamily: "FiraCode Nerd Font",
	rawSize: 12pt,

	linkColorForeground: aqua,

	bigSize: 150pt,
	bigColorForeground: gradient.conic(..color.map.rainbow),
)

#let themes = (
    dark: theme,

    light: {
        theme.insert("baseColorBackground", white)
        theme.insert("baseColorForeground", black)

        theme
    },
)

#let slides(
    title: "",
    subtitle: "",
    shortTitle: "",
    author: "",
    org: "",
    date: datetime.today().display("[year repr:full]-[month repr:numerical padding:zero]-[day padding:zero]"),
    theme: themes.dark,
    logo: none,

    body,
) = {
// Set PDF metadata.
set document(
    title: title,
    author: author,
)


// Set foundational page and text properties.
set text(
    fill: theme.baseColorForeground,
    font: theme.baseFontFamily,
    size: theme.baseSize,
)

let section = state("section", none)

set page(
    paper: theme.paper,
    fill: theme.baseColorBackground,

    header: [
        #box(width: 100%)[
            #set align(horizon + left)
            #set text(size: 10pt)

            #if logo != none [
                #box[#image(logo, width: 3cm)]
            ]

            #set align(horizon + right)
            #box[#section.display()]
        ]
    ],

    footer: [
        #set text(theme.baseSizeFooter)
        #box[#author #if shortTitle != "" [\/ #shortTitle] #h(1fr) #counter(page).display("1")]
    ],
)

show link: it => {
    set text(fill: theme.linkColorForeground)
    underline(it)
}

show raw: it => {
    set text(
        font: theme.rawFontFamily,
        size: theme.rawSize,
    )

    it
}

show heading.where(level: 1): it => {
    section.update(it.body)
    pagebreak()

    set align(horizon + center)

    text(
        size: theme.baseSizeHeading1,

        it.body,
    )
}

show heading.where(level: 2): it => {
    pagebreak()

    set align(top + left)

    text(
        size: theme.baseSizeHeading2,
        weight: "medium",

        emph(it.body),
    )

    v(1pt)
}

let titleSlide() = {
    set align(horizon + center)
    set page(footer: none, header: none)

    text(
        fill: theme.titleColorForeground,
        size: theme.titleSize,
    )[#title]

    linebreak()
    v(0.5cm)

    text(size: theme.titleSizeSubtitle)[#emph(subtitle)]

    linebreak()
    v(0.5cm)

    text(size: theme.titleSizeMetadata)[#author | #date]
}

show <big>: it => {
  set align(horizon + center)
  //set page(header: none, footer: none)
  set text(
      size: theme.bigSize,
      fill: theme.bigColorForeground,
  )

  strong(it)
}

titleSlide()

body
}
