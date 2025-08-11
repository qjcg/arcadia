Go Scratch
That Itch

John Gosset
Twitter: @jgosset_
GitHub: @qjcg

@img/avatar.jpg

- Consultant & software developer in Montreal, Canada
- Server software and embedded systems
- Organizations large & small
- RedHat, Software Carpentry
- Believe programming is for everybody.

What?
- "There should be code that does that!"
- Here to convince you to do, instead of not do
- Your problems matter

How?
- Small is different
- Scratching itches
- Some well-known itches
- Why Go is well-suited to
  programming in-the-small

# Small is different

@img/cat2.png

Stephen J. Gould:
Size & Shape

Size & Shape:
- As you scale up software, surface area (ex: public
  API) can be dragged down by supporting utility
  functions, tooling, boilerplate code, etc
- Avoid unnecessary complexity
- Ask: "Can we keep this small?"

In-the-large VS
In-the-small

IN-THE-SMALL
one thing well
often single developer
complete in itself

IN-THE-LARGE
industrial teams
months to years
submodules

Every good work of software starts by scratching a
developer's personal itch.
    -- Eric S. Raymond

THE STORY
 OF UNIX

@img/space_travel.png

Go Scratch
That Itch!

Jennifer
Dewalt

Go Scratch
That Itch!

180 websites in
180 days

A personal example

Go Scratch
That Itch!

     HOREB
י 𐅓 𝝖 🍫 ➩ ◼ 𐄨
⚈ 🂶 𐅾 ✷ 𐇒 ✃ ⚡ ℺
𐅪 ❿ 𐄳 ◮ ℌ 𐄫 ℍ ✚
🁖 ₹ ׃ ⌂ 𐇬 𝒔 ⛺

Don't like it? Cool with me!
      Your opinion is
     **not the point!**

Scratching Itches:
Everybody has problems

- Your problems are important!
- You have **unique local knowledge**
- Contrast with design by committee

Principles for programming
in the small

- Solve the problem in the concrete first
- Consider existing tools
- The [Unix Philosophy](https://en.wikipedia.org/wiki/Unix_philosophy)
- YAGNI
- Make your objectives clear (see eg.: [README-driven development](http://tom.preston-werner.com/2010/08/23/readme-driven-development.html))

---

### Unix Philosophy

> This is the Unix philosophy: Write programs that do one thing and do it well.
> Write programs to work together. Write programs to handle text streams, because
> that is a universal interface. -- Doug McIlroy

- [Rob Pike's Five Rules of Programming](http://users.ece.utexas.edu/~adnan/pike.html)
- [Plan9's](https://en.wikipedia.org/wiki/Plan_9_from_Bell_Labs) [Acme], an
  *"Integrating"* Development Environment
    - Watch [Russ Cox's Tour of Acme screencast](https://www.youtube.com/watch?v=dP1xVpMPn8M)

[Acme]: https://en.wikipedia.org/wiki/Acme_(text_editor)

---

### YAGNI

You Ain't Gonna Need It

<img class="slimg" height=400 src="img/yagni.png" />

---

# Go is good for programming in the small

---

### Go Spec

<img height=400px src="img/gomoving2.png" />

- Consider the name "Go": short & simple, moving & doing
- [The Go Spec](https://golang.org/ref/spec) is short and readable

---

### Go Says "No"

- [RSC on Generics](http://research.swtch.com/generic)
- [Go FAQ: Exceptions](https://golang.org/doc/faq#exceptions)
- Go is implemented as a procedural language (not e.g. primarily
functional)

---

### Easy Deployment

- Simple, minimalistic tooling (cf. Makefiles, venvs, package.json...)
- Cross compilation is trivial (`GOOS`, `GOARCH`)
- Static binaries

---

### Polymorphism Without Pain

- [Interfaces](https://gobyexample.com/interfaces) allow code re-use like with
  OO-style polymorphism, but via composition instead of requiring rigid object
  hierarchies

---

### Go is a shop-built jig

<img class="slimg" src="img/jig.jpg" />

- [Blog post by Rob Napier](http://robnapier.net/go-is-a-shop-built-jig)

> Go feels under-engineered because it only solves real problems.

---

### Go is a shop-built jig (cont'd)

> [...] little pieces of wood that help you hold plywood while you cut it, or spacers
> that tell you where to put the guide bar for a specific tool, or hold-downs that
> keep a board in place while you’re working on it. They’re not always pretty.
> They often solve hyper-specific problems and work only with your specific tools.
> And when you look at ones that have been used a lot, they sometimes seem a
> little weird.

---

### Go is a shop-built jig (cont'd)

> There might be a random cutout in the middle. Or some little piece that sticks
> off at an angle. Or the corner might be missing a piece. And when you compare
> them to “real” tools, “general” tools like you’d buy from a catalog, they’re
> pretty homey or homely depending on how you’re thinking about it.
> [...] really good at solving [...] very common problems for people who need
> to ship software

---

### Key Takeaways

Principles when programming in the small:
- Write down what you want (eg: [README driven development](http://tom.preston-werner.com/2010/08/23/readme-driven-development.html))
- Solve the problem in the concrete first
- Keep the [Unix Philosophy](https://en.wikipedia.org/wiki/Unix_philosophy) in mind
- YAGNI
- When in doubt, use brute force

---

### Key Takeaways (cont'd)

Go features good for programming in the small:
- small spec
- says "no" to keep the language small
- easy deployment
- interfaces for polymorphic code reuse without rigid object hierarchies
- excellent ecosystem and community

---

### Conclusion

- Programming in the small is good
- Go is good for programming in the small
- **Your problems matter**!
- You should...

--- gsti

# Go Scratch That Itch!

---

### References

- Background
    - [README Driven Development](http://tom.preston-werner.com/2010/08/23/readme-driven-development.html)
    - [Programming in the large and programming in the small]
    - [Unix Philosophy](https://en.wikipedia.org/wiki/Unix_philosophy)
    - [Worse is Better](https://en.wikipedia.org/wiki/Worse_is_better)
    - [When in doubt, use brute force](http://www.catb.org/jargon/html/B/brute-force.html)
    - [Stephen J Gould: Size and Shape](http://www.naturalhistorymag.com/editors_pick/1974_01_pick.html)
    - [The Art of Unix Programming: The Right Size of Software](http://www.catb.org/esr/writings/taoup/html/ch13s04.html)

[Programming in the large and programming in the small]: https://en.wikipedia.org/wiki/Programming_in_the_large_and_programming_in_the_small

---

### References (cont'd)

- Itches
    - [The Story of Unix & Space Travel](https://www.bell-labs.com/usr/dmr/www/spacetravel.html)
    - [Jennifer Dewalt: 180 Websites in 180 days](https://jenniferdewalt.com/)
    - [horeb](https://github.com/qjcg/horeb)
- Go
    - [The Go Spec](https://golang.org/ref/spec)
    - [Go Proverbs](https://go-proverbs.github.io/)
    - [Go is a shop-built jig](http://robnapier.net/go-is-a-shop-built-jig)
