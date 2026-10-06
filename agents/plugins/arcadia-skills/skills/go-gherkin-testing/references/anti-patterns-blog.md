---
description: A transcript and interpretation of a Cucumber Ltd podcast where Aslak Hellesøy, Matt Wynne, and Steve Tooke discuss common Cucumber anti-patterns — bad collaboration, poor living documentation, beginner mistakes, and more.
source: https://www.thinkcode.se/blog/2016/06/22/cucumber-antipatterns
---

# Cucumber Anti-Patterns

*By Thomas Sundberg — 2016-06-22*

This a transcript and interpretation of a podcast where [Aslak Hellesøy](https://twitter.com/aslak_hellesoy),
[Matt Wynne](https://twitter.com/mattwynne), and
[Steve Tooke](https://twitter.com/tooky)
from [Cucumber Ltd](https://cucumber.io/#company) have
a chat about the [BDD](https://en.wikipedia.org/wiki/Behavior-driven_development)
tool [Cucumber](https://cucumber.io/) and
anti-patterns they have come across.

The original podcast is available at
[https://cucumber.io/blog/2016/05/09/cucumber-antipatterns](https://cucumber.io/blog/2016/05/09/cucumber-antipatterns).

Cucumber is designed to be a collaboration tool that creates a living documentation, that is possible to automate,
of the behaviour you want in a system. You describe the wanted behaviour using a language called
[Gherkin](https://cucumber.io/docs/reference#gherkin).
If you are unfamiliar with Gherkin, it is probably a good idea to read up on it before you continue.

## Bad collaboration

Cucumber is a collaboration tool and should be used for collaborating when driving the implementation of software.

### When do you write feature files

One anti-pattern is that you write the feature files after you have written the code. This could arguably be the
worst anti-pattern when using Gherkin. It is very common when you start using Cucumber and BDD.

What you really want is to write the Gherkin before you write the software. This will allow you to actually
use the scenarios to drive the development and not the other way around, to document what you developed.

Why should you write the Gherkin before you write the software? Documenting the wanted behaviour is a way to make
sure that everyone agrees on what the software is supposed to do. When you have reached that agreement, you have a
common understanding of the problem, then it is a good time to start the implementation and write the code needed.
Writing code before you have come to a common understanding is one way to make sure you implement something that you
will have to change later.

Cucumber is a testing tool that allows you to test your assumptions and understanding of the problem before you
actually write the code to solve the problem. It will make it clear if the people involved agree on the wanted
behavior or if there is ambiguity that hasn't been discovered yet. Concrete examples before the implementation will
make it clear that everyone has the same understanding.

A product owner or business analyst, a dev, and a tester will all have different perspectives on the problem.
Understanding these different perspectives before writing the code increases the chance of creating something that
actually is what the end user needs.

### Business people create scenarios in isolation

One person, maybe the product owner or business analyst, writes the scenarios alone.
The result will be Gherkin that doesn't represent everybody's understanding. The scenarios written by
the product owner or business analyst alone tend to be un-testable.

To be able to automate the scenarios, they will have to be changed. This leads to Gherkin that isn't the Gherkin the
product owner or business analyst originally wrote and therefore it doesn't represent what they wanted. They will
lose interest and the Gherkin will be something devs or testers use instead.

The common understanding is lost.

There is also a chance that when the devs start to reformulate the Gherkin, they misunderstand what the product
owner or business analyst actually meant and ends up with something that is incorrect.

### Devs or testers writing scenarios without talking to business people

This is similar to the case where product owner or business analyst writes the scenarios alone. But in this case the
examples tend to be unrealistic. Sometimes really dry. The examples are not talking about real users,
[personas](https://en.wikipedia.org/wiki/Persona_%28user_experience%29), but instead talking about user1 and user2.

In this case the developers missed out on an opportunity to have realistic users, realistic data and ended up with
something really boring and dry.

### Too high level

If a scenario is on a too high level, you can't really trust it.
You can probably not tell what it actually will do because it is too high level and vague.

An example could be:

```gherkin
Given a bank account
When I withdraw some money
Then the balance should be the original balance minus the amount withdrawn
```

This is an example that expresses a business rule. It doesn't contain any concrete example. We don't know the
original balance; we don't know how much that was withdrawn so we can't know the outcome. It is not concrete.

Another example is:

```gherkin
Given the system is turned on
When it is used
Then it should work perfectly
```

This leaves a lot of freedom to the implementer. It also trusts that the implementer has understood the requirement
perfectly and will not make any mistake. It is totally useless as an example that drives the development.

This is a trap that it is easier to fall into if the product owner or business analyst writes the scenarios alone.
With more people writing scenarios, more perspectives are taken into account. Vague examples can be made
concrete, unclear examples can be questioned, and unclear rules can be found.

Finding the right balance between too vague and too many details is hard. This is something experience will
teach you.

## No living documentation

Cucumber is a documentation tool and can be used to create a good, living, documentation about a system.

### When you read Gherkin and it is bad documentation

The litmus test could be: Take a scenario, show it to someone who isn't familiar with the domain and ask that person
if they can describe the functionality the scenario represents. If that person can't describe what the system is
supposed to do, then you have a case of bad documentation.

An example can be a long scenario, ten or more steps, and lots of [incidental details](https://www.vocabulary.com/definition/incidental) that don't describe an example
of a business rule but rather a journey through the application. How it should be used rather than the
behavior we are trying to verify.

### Incidental details

Someone tried to tell a good story but instead overload the reader with too much information.
The incidental details are often there because the author has written a test rather than a documentation of the
desired behavior.

An example could be:

```gherkin
When I sign up as Matt, my password is password, my password confirmation is password, and I check my bank balance
Then I find $100
```

The purpose with the scenario is to check the bank balance. In this case, it isn't relevant what your password
is or how you login. It is there because the author needed the values when automating a test.

The problem with these incidental details is that they obscure the essence of the scenario. It is hard to understand
what we really try to test. In this case, the reader would have to guess. Are we testing the password functionality
or are we testing the bank balance? Or are we testing both in one scenario?
All the details about passwords and bank balance makes it unclear what the purpose of the scenario is.

### Hard to tell what you are testing

Are you testing one rule or are you testing several rules in the same scenario?

There are two problems here:

- Testing multiple things at the same time
- Not creating a clear documentation for checking if this is the desired behaviour

You may be forced to read between the lines to find out what the actual essence of the scenario is. If you are
unlucky, the scenario might not have an essence. It is unfocused and aims at everything at the same time.

### Bad name on a scenario

A good name on a scenario tells the reader what this is about. No name leaves the reader guessing.

A name that tries to summarize the entire behaviour of the scenario, such as
`Scenario: Sign up, login, go to balance screen, check balance, logout`
is boring and you will lose the interest of the reader.

A good place to sum up the essence of a scenario is in its name. A name could be `Scenario: Check balance`.

The name indicates that any details about something else are incidental. Any details about a password are
incidental and should be removed. If a password is needed for the automation, it can be added in
the automation layer.

A good naming algorithm could be to use the *Friends* metaphor: `This is the one where ...`

Examples:

- This is the one where the balance is positive
- This is the one where I have an overdraft on my account
- This is the one where I change my password

The behavior is then described in the `Given/When/Then` of the scenario.

### Not using the narrative section of a feature

The first part in a [feature file](https://cucumber.io/docs/reference) is called the *narrative section*.
You can write whatever you want here. You can leave it blank or add useful things.

Leaving the narrative section blank or adding a user story like the one below are two ways of applying an
anti-pattern:

```gherkin
As a user,
I want to check my balance,
so I know what my balance is
```

This user story doesn't give you any extra information and is therefore pretty useless.
Someone just filled out a template without really reflecting over why someone else wants this behavior.

Instead, put useful information for the reader in the beginning of the feature. Describe the business rules;
the abstract rules that we will describe examples for.

An example of a rule could be: *After you have done a withdrawal, that withdrawal should show up in your list of
transactions*.

The format you describe the business rule in is not that important. Describe them using a bullet list if that's better.
The examples that follow will make them concrete.

The narrative section is the place to put questions or uncertainties in as well.

Think of the feature file as living documentation where you put the rules discovered and unanswered questions, known
unknowns. This is our current understanding of this feature. There is stuff we know and there is stuff we don't know yet.

## Beginners mistakes

There are some mistakes commonly done by beginners.

### Lots of user interface details

The system is used through its user interface. This sometimes leads to scenarios that talks about going to a
specific url, click a specific link, find an element using a specific css selector etc. It doesn't really tell us what
the purpose of using the program is. It is hard to understand why the user wants to do this journey.

One problem with this is understanding the purpose. Another problem is that user interfaces changes a lot more
frequently than the underlying domain logic.

UI trends change more often than the actual business. This will break your tests even if the business
hasn't changed. UI tests are very slow and brittle. It takes a long time to fire up a UI testing tool.

Testing through the UI will make you stuck at the top of the [testing pyramid](http://martinfowler.com/bliki/TestPyramid.html).
All of your testing will be done through the user interface. This is a bad and painful place to be since it is slow and brittle.

Scenarios littered with UI details will make it impossible to test anything below the user interface.

Too many user interface details is also very poor documentation. It is hard to understand the business rule by
understanding how the navigation through a system works. The navigation through a program doesn't tell us why we are
using the program.

The problem you solve by using this program is not described using words from the domain, it is described using the
generic terms of user interfaces. Not the domain of your business. You miss the nouns and verbs that describe your
business problem that you would like to use deeper down in your code.

This anti-pattern usually comes from the fact that you wrote the scenarios after you actually implemented the
solution.

### Describing actions using the personal pronoun *I*

Most systems have behavior that is used by multiple users or actors.
This means that you want to talk about a specific user, a *persona*, if you can.
A persona will give you context about what the system should be able to do to support a specific category of users.

If a system supports a university, there is probably a difference between a 21 year old student and a 53 year old
administrator. The role, student or administrator, gives you different contexts.

### Documenting boring scenarios

Some scenarios are very boring and perhaps just needed in the beginning of a project. An example is:

```gherkin
Given my bank account is empty
When I check my bank balance
Then it should be 0
```

You probably need certain scenarios in the beginning of the project or when you start developing a certain feature.
Some of them will, however, be obvious after a while and covered by other tests.

### Keeping all scenarios forever

Not all scenarios will bring value forever. You may delete scenarios after a while if you are
certain that this case is covered in other tests. But before you delete them, make sure that the behavior is covered
somewhere else so you don't lose it.

Instead of deleting scenarios, you can rewrite them to document something more interesting. There are perhaps edge
cases that can be covered with a rephrased scenario.

Keeping scenarios because they were written once is not a good argument. Don't do that.

### No clear separation between Given/When/Then

It happens that beginners have a hard time grasping the difference between `Given`,
`When` and `Then`. The problem is that from a technical perspective, there is no difference, Cucumber will
treat them all equal.

What is the difference then?

- `Given` is the context — the past
- `When` is an action that changes the system — the present
- `Then` is the expected outcome — the near future

A metaphor could be: Going to the theater.

You are sitting in your chair and the curtain is drawn. Behind the curtain there is a lot of stage workers
preparing. Putting things in place. This is the `Given`. This is where you do the setting up of your system,
creating initial objects, prepare the database, navigating to a specific page and so on.

The curtain is drawn apart and the act starts. Things start to happen, one thing leads to another. Actors start
doing things and they say their lines. This is the `When`. This is triggering the important event that we want to
happen in a specific context.

One thing on stage leads to another thing, the expected outcome. What happened in the story as a result of what an
actor did. This is the `Then`. The observable change that we want to assert.

- The past — the preparation is `Given`
- The present — the action is the `When`
- The near future — the expected outcome is the `Then`

### Multiple Whens

That is a `When` is followed by one or many `And`s.

There are cases when it can be ok. Say that two different persons does two different things in a `When`
step, then it can be ok. This is an example of similar events.

When there are two completely different events, then it may be a bad idea to have an
implicit `When`, expressed as an `And`. It can be an indicator that you should split this scenario into two
different scenarios. Chances are that you describe two different behaviors in the same scenario.

## Double edged sword

There are cases that may or may not be anti-patterns. They can be used in a bad way, but they also enable you to do
really good things. They are like a double edged sword and depends on the context.

### Scenario outline

Scenario outline is an example of something that easily can be overused and lead to slow tests.
This is especially true if they contain a lot of UI details.

It is very easy to add a lot of scenarios when you use an outline. If the scenarios are slow, you will end up with a
really slow test suite.

Using scenario outlines to verify complicated algorithms can be a valid use case. But probably not through
the slow user interface.

An example can be to verify a complicated insurance algorithm where there are lots of data that will
create different outcomes. In this case, it is important that you can reach the algorithm without
going through the UI. This requires the system to be built so it is testable.

If the scenarios are fast, then it may be ok to use Scenario outlines. If they are slow, do not use scenario
outlines.

### Multiple Thens in the same scenario

This is not necessarily an anti-pattern. An example could be when something is returned in a retail system:

- The customer should be refunded
- The inventory should be updated
- The finance system should be updated
- A message should be sent to a warehouse that the returned item needs to be picked up

These are all connected. Verifying them in one go could be a good idea. At the same time, these outcomes are
the result of four different business rules. Splitting them into four different scenarios will make the business rule
much more explicit and will probably help you understand the system better. Splitting these will also allow the
development team to implement the rules incrementally; they don't have to be implemented at the same time.
This only applies to independent outcomes.

There are examples when it is not possible to split a scenario. Withdrawing money from a bank account is such an
example:

- The customer should get their money from an ATM machine, say £50
- The customer account should be debited £50

These outcomes are depending on each other and can't be divided into different scenarios. The bank will not
appreciate if the customers gets money without the account being debited. The other way around is probably also
true — most customers will not be happy if their account is debited but they didn't get any cash.

## Holiday photos

The features can be seen as a reminder of what you talked about and agreed on in the session when you wrote them.
They should contain enough details for you to remember the most important details.

They should serve the same purpose as a holiday photo — remind you of a great time.

## Conclusion

There are many mistakes you can do when you use Cucumber and implement BDD. It is not the end of the world if
you do them, but it is better to avoid them. You are now equipped with some knowledge about some of the
anti-patterns that are possible to do.

## Resources

- [A podcast from Cucumber Ltd](https://cucumber.io/blog/2016/05/09/cucumber-antipatterns) with Aslak Hellesøy, Matt Wynne, and Steve Tooke
- [Cucumber](https://cucumber.io/) — A BDD tool for simple, human collaboration
- [Thomas Sundberg](http://www.thinkcode.se/blog/about) — your listener and author
