---
name: specs
description: Help create software specifications from rough notes
---

# Specs - Software Specification Generator

## When to use this skill

Use this skill whenever the user wants to create a software specification document from rough notes, brainstorming, or a brain dump about software to be written.

## Overview

This skill guides the user through a structured process to transform
rough ideas into a formal software specification. It asks targeted
follow-up questions to flesh out details and then produces a
well-organized spec document.

## Supported specification types

This skill can create the following types of specifications (user may select one or more):

- **Elevator Pitch** - A concise, one-sentence product vision based on the Scrum.org template. Good for:
  - Defining clear product focus
  - Communication with stakeholders
  - Marketing and product visioning
- **Tagline** - A short, pithy, and fun phrase that captures the essence of the project. Good for:
  - Branding
  - Quick introductions
  - Slack/GitHub project descriptions
- **PR-FAQ (Press Release and Frequently Asked Questions)** - Amazon-style product visioning document. Good for:
  - Thinking big and starting from the customer experience
  - Working backwards from the launch/announcement
  - Aligning teams on product vision and specific details
- **EARS (Easy Approach to Requirements Syntax)** - A structured format for writing clear, unambiguous requirements. Good for:
  - systems requiring formal requirements
  - Complex business rules
- **README.md** - A simple, concise overview of what the tool is and a quickstart.
  - Quick introduction
  - Installation instructions
  - Basic usage (Quickstart)
- **SRS (Software Requirements Specification)** - IEEE 830-style specification. Good for:
  - Traditional software projects
  - Customer-facing documentation
  - Regulatory compliance

- **PRD (Product Requirements Document)** - Product-focused specification. Good for:
  - Product managers and business stakeholders
  - Startups and fast-moving teams
  - Market-facing products

- **User Stories** - Agile/Scrum format. Good for:
  - Sprint planning
  - Developer handoffs
  - Backlog management

- **Technical Design Document (TDD)** - Technical architecture and implementation plan. Good for:
  - Complex systems
  - Team technical discussions
  - Long-term maintenance

## Process

### 1. Initial input gathering

When the user provides rough notes, ask for any missing context in these areas:

- **Purpose**: What problem does this software solve? Who is it for?
- **Scope**: What are the boundaries? What is explicitly out of scope?
- **Current state**: Is this new software, a rewrite, or an extension? What exists now?
- **Success criteria**: How will we know it's done?

### 2. Default specification documents

By default, after initial context gathering, generate the following documents:

- `README.md` (at project root)
- `docs/PITCH.md` (Elevator Pitch)
- `docs/TAGLINE.md` (Tagline)
- `docs/PRD.md` (Product Requirements Document)
- `docs/PR-FAQ.md` (Press Release and FAQ)

Ask the user if they want additional types from the [supported list](#supported-specification-types).

### 3. Follow-up questioning

**Process for defaults** (gather in sequence):

1. **Elevator Pitch (PITCH.md)**: Ask about target customer, specific need, product category, key benefit, primary competitor.
2. **Tagline (TAGLINE.md)**: Ask about core "vibe" or personality. Generate 20 possible taglines (&lt;10 words each) as numbered list based on all gathered info. Ask user: "Pick one by number, suggest modification, or new idea?" Wait for selection before proceeding.
3. **PRD (PRD.md)**: Ask about user personas, market context, success metrics, prioritization.
4. **PR-FAQ (PR-FAQ.md)**: Press Release (headline, problem, solution, quotes); FAQs (external/internal concerns).
5. **README.md**: Installation steps, basic usage examples.

**Then for additional specs** (if requested):

**For Elevator Pitch**: Ask about target customer, specific need, product category, key benefit, and primary competitor.
**For Tagline**: Ask about the core "vibe" or personality of the project.
**For PR-FAQ**:
- **Press Release**: Ask about the headline, target customer, the problem, the solution, a hypothetical leadership quote, and a hypothetical customer quote.
- **FAQs**: Ask about external customer concerns (price, ease of use, security) and internal stakeholder concerns (market size, technical complexity, success metrics).
**For EARS**: Ask about system states, events, conditions, and response requirements
**For README.md**: Ask about installation steps and basic usage examples.
**For SRS**: Ask about functional/non-functional requirements, interfaces, constraints
**For PRD**: Ask about user personas, market context, success metrics, prioritization
**For User Stories**: Ask about user roles, acceptance criteria, and value
**For TDD**: Ask about architecture choices, technology stack, data models, APIs

### 4. Document generation

After gathering sufficient information, generate the specification document(s) in the requested format(s).

**For the Elevator Pitch**, follow this exact format:
> For [target customer] who [statement of the need], the [product name] is a [product category] that [key benefit, compelling reason to buy]. Unlike [primary competitive alternative], our product [statement of primary differentiation].

**For the Tagline**, use the user-selected short, pithy phrase (&lt;10 words) from the 20 options (or modification). Reference in other docs where appropriate (e.g., README header).


**For the PR-FAQ**, use the following structure (Note: In the final document, the PR sections should flow naturally as standard press release paragraphs, NOT as explicitly labeled sections like "The Problem" or "Leadership Quote"):
1. **Press Release (PR)**
   - **Headline**: Catchy name and main benefit
   - **Sub-headline**: More detail on target audience and benefit
   - **Body Paragraphs**:
     - **Summary**: Describe the announcement/launch.
     - **Problem**: Describe the pain point being addressed.
     - **Solution**: Explain how the product solves it.
     - **Leadership Quote**: Insights from a spokesperson.
     - **Customer Quote**: Testimonial from a satisfied user.
     - **Closing**: How to get started and call to action.
2. **Frequently Asked Questions (FAQ)**
   - **External FAQ**: Questions a customer might ask (e.g., "How much does it cost?", "How do I install it?")
   - **Internal FAQ**: Questions an internal stakeholder might ask (e.g., "What is the market size?", "How will we measure success?")

Structure the output clearly with appropriate sections for each spec type.

## Generating specifications

When generating specifications:

- Use clear, concise language
- Organize with appropriate headings and subheadings
- Include all gathered information in structured form
- Add placeholder sections for areas where information is incomplete
- Mark any assumptions or uncertainties clearly
- For multiple spec types, either generate separate documents or a combined document with clear section divisions

## Output format

Present the final specification as markdown that can be saved to a file.

**File Placement Rules:**
- `README.md`: Always place at the top-level of the Go module.
- All other documents (e.g., `PITCH.md`, `PRD.md`, `TAGLINE.md`, `SRS.md`, `requirements.md`): Always place under a `docs/` directory relative to the module root.

Suggest an appropriate filename based on the project name.

## Example interaction flow

User: "I want to build a task manager app..."

You (Specwriter):
1. Ask: "What's the purpose of this task manager? Who are the target users?"
2. "Great! By default I'll generate README.md (root), docs/PITCH.md, docs/TAGLINE.md (pick from 20), docs/PRD.md, docs/PR-FAQ.md. Additional types?"
3. Gather info: Elevator Pitch details → Generate/pick tagline → PRD/PR-FAQ/README details
4. Generate all documents using selected tagline
