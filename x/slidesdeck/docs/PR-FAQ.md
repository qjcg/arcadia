# PR-FAQ: Slidesdeck - The Plain Text Slideshow Creator

## 1. Press Release (PR)

**FOR IMMEDIATE RELEASE**

### **Slidesdeck Launches: Simplifying Professional Presentations for the Markdown and Org-mode Era**

**A single-binary CLI tool that transforms plain text into interactive, self-contained HTML slideshows with zero external dependencies and built-in break management.**

**ALDIE, VA – March 5, 2026** – Today, we are announcing the official release of Slidesdeck, a command-line tool designed for developers and technical writers who want to create high-quality presentations without leaving their text editors. Slidesdeck converts Markdown and Emacs Org-mode files into lightweight, polished HTML5 slideshows that run anywhere, ensuring that your content remains the focus of every presentation.

For too long, technical presenters have been forced to choose between two extremes: heavy, proprietary GUI software like PowerPoint that breaks developer workflows, or complex, multi-dependency CLI tools like Pandoc that require extensive configuration and external templates. These solutions often result in "dependency hell" or fragile presentation files that fail to render correctly without a stable internet connection, often failing at the most critical moments during a live talk.

Slidesdeck offers a "working backwards" approach to presentations by focusing on the text first, allowing authors to stay in their creative flow. With a single command, Slidesdeck produces a single, self-contained HTML file containing all necessary styles and interactivity. It leverages modern web technologies like Tailwind CSS, daisyUI, and Alpine.js to deliver a beautiful, interactive experience that includes a real-time theme switcher and a unique "Pause Mode" for managing presentation breaks.

"We built Slidesdeck because we were tired of fighting with our presentation tools ten minutes before a talk," said the lead architect. "We wanted something that respected the plain-text philosophy of developers but didn't sacrifice the 'vibe' of a modern web application. Slidesdeck is about speed, portability, and staying in your editor."

"Slidesdeck has completely changed how I prepare for meetups," said a senior platform engineer. "I can take my existing project notes in Org-mode, add a few separators, and have a professional deck ready in seconds. The fact that it's just one HTML file means I never have to worry about missing assets or broken styles ever again."

Users can install Slidesdeck today via Go: `go install github.com/charmbracelet/arcadia/cmd/slidesdeck@latest`. To create your first deck, simply run `slidesdeck my-notes.md`. For more information and documentation, visit our official repository.

---

## 2. Frequently Asked Questions (FAQ)

### **External FAQ (Customer-Facing)**

**Q: What file formats does Slidesdeck support?**
A: Slidesdeck natively supports Markdown (.md) and Emacs Org-mode (.org) files.

**Q: Do I need to host the generated HTML on a web server?**
A: No. The output is a completely self-contained HTML file. You can open it directly from your disk in any modern web browser, and it will work perfectly offline.

**Q: Can I change the look and feel of my slides?**
A: Yes! Slidesdeck bundles the entire daisyUI theme library. You can set a default theme with the `--theme` flag or switch themes at runtime using the "Command Palette" by pressing `t`.

**Q: Can I search within my presentation?**
A: Yes. By pressing `/`, you can access a dedicated Search Command Palette that allows you to quickly find and jump to any slide by its title or content.

**Q: How do I manage breaks during long presentations?**
A: Slidesdeck features a built-in "Pause Mode." By pressing `Shift+P`, you can set a countdown timer or a target end-time (e.g., "Back at 2:00 PM") with a custom message. This persists even if you accidentally close your browser.

**Q: Is there syntax highlighting for my code snippets?**
A: Absolutely. Slidesdeck uses the Chroma library to provide professional, high-quality syntax highlighting for hundreds of programming languages, with line numbers enabled by default.

### **Internal FAQ (Stakeholder-Facing)**

**Q: Why choose Go for this tool instead of Node.js?**
A: Performance and portability. Using Go allows us to distribute a single, statically-linked binary that converts even large decks in milliseconds. It eliminates the need for users to manage a `node_modules` folder or a runtime environment.

**Q: How do we measure the success of Slidesdeck?**
A: Success is measured by user adoption within the developer community, specifically looking at GitHub stars, community-contributed themes, and its use in major technical conferences.

**Q: How does the tool handle complex Org-mode features?**
A: We use the `go-org` library, which provides a robust parser for the Org-mode specification, ensuring that lists, tables, and source blocks are rendered accurately.

**Q: How do we maintain the CSS if daisyUI has so many themes?**
A: We've built a modular asset pipeline using `esbuild`. Each theme is stored separately and bundled into the binary at build time, keeping the development process clean and the final binary optimized.

**Q: What is the plan for future "Version 2.0" features?**
A: Our roadmap includes a "Watch Mode" for live preview during editing and the ability to automatically embed local images as Base64 strings for even greater portability.
