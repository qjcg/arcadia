#import "@preview/codelst:2.0.0": sourcecode
#import "@preview/tiaoma:0.2.0"

#import "template.typ": slides, themes

#show: slides.with(
    title: "No Code\nLow Code\nGo Code",
	subtitle: "Embracing the code that isn't there",
    shortTitle: "No Code, Low Code, Go Code",
	author: "John Gosset",
	theme: themes.dark,
)


= KISS

== Knuth vs. McIlroy

- #link("https://dl.acm.org/doi/10.1145/5948.315654")[Programming Pearls: A Literate Program] (1986)
- #quote()[
Read a file of text, determine the n most frequently used
words, and print out a sorted list of those words along with
their frequencies.
]

#align(center)[
  #image("./imgs/knuth-vs-mcilroy.png", height: 50%)
]

== McIlroy's Low Code Solution

#sourcecode(frame: none)[
```sh
tr -cs A-Za-z '\n' |
tr A-Z a-z |
sort |
uniq -c |
sort -rn |
sed ${1}q
```
]

#quote()[
Knuth has shown us here how to program intelligibly, but not wisely. I buy the discipline. I do not buy the result. He has fashioned a sort of industrial-strength Fabergé egg—intricate, wonderfully worked, refined beyond all ordinary desires, a museum piece from the start.
]

== Keep it Simple, Superhero

#set align(center)

#box(width: 80%)[
  #set text(size: 16pt)
  #set align(horizon + center)

  #set quote(block: true)

  #quote(attribution: [Ken Thompson])[
  One of my most productive days was throwing away 1000 lines of code.
  ]

  #quote(attribution: [Gordon Bell])[
  The cheapest, fastest, and most reliable components are those that aren't there.
  ]

  #quote(attribution: [Albert Einstein])[
  Everything should be made as simple as possible, but not simpler.
  ]

  #quote(attribution: [Steve Jobs])[
  #strike[Simplify, Simplify,] Simplify
  ]
]

#set align(top + left)

== Go was born to simplify

#quote(attribution: [Rob Pike, Less is Exponentially More])[
Back around September 2007, I was
doing some minor but central work on an enormous Google C++ program,
one you've all interacted with, and my compilations were taking about
45 minutes on our huge distributed compile cluster. ... At this point
I asked myself a question: Did the C++ committee really believe that
was wrong with C++ was that it didn't have enough features? ... I
started another compilation, turned my chair around to face Robert,
and started asking pointed questions. Before the compilation was done,
we'd roped Ken in and had decided to do something. We did not want to
be writing in C++ forever...
]

== Kent Beck: 4 Rules of Simple Design

#align(horizon + center)[
  #image("./imgs/beck-design-rules.png", height: 50%)
]

= Low Code?

== Definitions: IBM

/ Low Code:

  "Low-code is a visual approach to software development that enables faster delivery of applications through minimal hand-coding." (IBM)

  (#link("https://www.ibm.com/topics/low-code")[Ref])

== Definitions: AWS

/ Low Code:

  "Low code or no code is a general phrase for development platforms and tools designed to help business users design and develop applications. Professional developers also use low code or no code tools for repetitive tasks that do not require conventional software development."

  (#link("https://aws.amazon.com/what-is/low-code/")[Ref])

== Definitions: No One Owns "Low Code"!


- There's lots of marketing copy by software vendors, but not "one true definition" of low-code
- The original low code... the function
- Can declarative configuration in YAML be "low-code?"
- Core features:
  - Code reuse
  - Working at the right level of abstraction
  - Out of the box functionality
  - Visual modelling (BUT is this a must?)
  - Empower "citizen developers"

== Low Code Tools

- SaaS offerings
  - AWS Workflow Studio
  - Microsoft Power Apps
- Open Source / Self-Hosted
  - Benthos (Go)
  - Telegraf (Go)
  - StackStorm
  - n8n.io
  - node-red
  - Vector

= Case Study: Benthos

== Benthos

https://www.benthos.dev

#align(horizon + center)[#image("./imgs/benthos.png", height: 70%)]

== Benthos: Basics

#set text(size: 22pt)

#sourcecode(frame: none)[
```yaml
# config.yaml
input:
  generate:
    mapping: root = "Hello, Benthos!"
    count: 1

pipeline:
  processors:
    - mapping: content().uppercase()

output:
  stdout: {}
```
]

#set text(size: 18pt)

== Benthos: HTTP

#sourcecode(frame: none)[
```yaml
# config.yaml
input:
  http_client:
    url: https://tradestie.com/api/v1/apps/reddit
    rate_limit: reddit

pipeline:
  processors:
    - unarchive:
        format: json_array

    - mapping: |
        root = this
        root.yolo = true
```
]

== Benthos: Unit Tests

#set text(size: 16pt)

#sourcecode(frame: none)[
```yaml
# config.yaml
input:
  generate:
    mapping: root = "Hello, Benthos!"
    count: 1

pipeline:
  processors:
    - mapping: content().uppercase()

tests:
  - name: Content must be uppercase
    target_processors: /pipeline/processors/0
    input_batch:
      - content: hello
    output_batches:
      - - content_equals: HELLO
```
]

== Benthos: Custom Plugins (in Go)

#sourcecode(frame: none)[
```go
func (r *reverseProcessor) Process(ctx context.Context, m *service.Message) (service.MessageBatch, error) {
	bytesContent, err := m.AsBytes()
	if err != nil {
		return nil, err
	}

	newBytes := make([]byte, len(bytesContent))
	for i, b := range bytesContent {
		newBytes[len(newBytes)-i-1] = b
	}

	if bytes.Equal(newBytes, bytesContent) {
		r.logger.Infof("Woah! This is like totally a palindrome: %s", bytesContent)
		r.countPalindromes.Incr(1)
	}

	m.SetBytes(newBytes)
	return []*service.Message{m}, nil
}
```
]

== Benthos: Studio

= In Closing

== Takeaways

- Keep it simple
  - Or, as Rob Pike put it: "Less is Exponentially More"
- Prioritize writing the interesting code that matters
- Consider low code for simple bridging, enrichment, filtering
- Benthos is a sweet spot for Go developers:
  - Low code stream processing
  - Tons of out of the box plugins
  - Awesome observability instrumentation
  - Can be extended with Go!


== Further Reading

- Knuth & McIlroy
  - #link("https://dl.acm.org/doi/10.1145/5948.315654")[Programming Pearls: A Literate Program]
  - #link("http://www.leancrew.com/all-this/2011/12/more-shell-less-egg/")[More Shell, Less Egg]
- Rob Pike: #link("https://commandcenter.blogspot.com/2012/06/less-is-exponentially-more.html")[Less is Exponentially More]
- Kent Beck: #link("https://martinfowler.com/bliki/BeckDesignRules.html")[4 Rules for Simple Design]
- Benthos
  - #link("https://www.benthos.dev")[Benthos]
  - #link("https://github.com/benthosdev/benthos-plugin-example")[Benthos Plugin example repo (Go)]
  - #link("https://studio.benthos.dev/")[Benthos Studio]
  - #link("https://www.youtube.com/watch?v=EA2Pco3EvpU")[Stream Processing the Easy Way (with NATS and Benthos)]

== Thank You!

#set align(horizon + center)

John Gosset

#link("https://www.linkedin.com/in/gossetj/")[LinkedIn] / #link("https://github.com/qjcg")[GitHub]

#link("https://github.com/qjcg/talks")[Slides + Code]
