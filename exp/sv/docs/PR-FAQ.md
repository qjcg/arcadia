# PR-FAQ: sv

## Press Release

### Headline
**sv: Finally, Seamless Semantic Versioning for Monorepos**

### Sub-headline
Stop fighting with global tags. Automate independent versioning for
every module in your monorepo with path-based tagging.

### Summary
Today we announce `sv`, a new CLI tool that brings the simplicity of
automated semantic versioning to the complex world of monorepos. `sv`
analyzes your Git history to calculate the next version of your
software, but unlike traditional tools, it understands that different
directories in your repo may need different versions.

### Problem
In a monorepo, a simple bug fix in a utility library shouldn't force a
version bump for the entire project. However, most semantic versioning
tools only look at the repository as a single unit. This leads to "tag
pollution," where developers are forced to use brittle scripts or
manual processes to manage independent module releases.

### Solution
`sv` solves this by introducing path-based tagging as a first-class
citizen. It automatically scopes its analysis to specific
directories. When you run `sv` inside a module, it only looks at the
tags and commits relevant to that module. It uses the familiar
Conventional Commits standard to decide if you need a major, minor, or
patch bump.

### Leadership Quote
"We loved the simplicity of tools like svu, but they broke down the
moment we moved to a monorepo," said the lead developer of
Arcadia. "sv gives us that 'it just works' experience while respecting
the boundaries of our modular architecture."

### Customer Quote
"Automating our releases used to be a nightmare of grep and custom
shell scripts. With sv, our CI/CD pipeline just asks 'what's next?'
and gets a perfectly scoped tag for every module in the repo."

### Closing
Get started today by installing `sv` via Go. Run `sv next` in any
subdirectory to see the future of your versioning.

## Frequently Asked Questions (FAQ)

### External FAQ
**How does sv know which directory belongs to which module?**

`sv` looks for boundary markers like `go.mod`. You can also explicitly
set the module path using the `--path` flag.

**Does this work with existing 'v' tags?**

Yes. If run at the root, it behaves exactly like traditional
versioning tools, looking for `v1.2.3` style tags.

### Internal FAQ
**How does this handle commits that touch multiple modules?**

If a commit touches multiple modules, `sv` will count that commit
toward the version calculation of *each* affected module. This ensures
that cross-cutting changes are reflected in all relevant version
bumps.

**Is it compatible with caarlos0/svu?**

Yes, we aim for CLI flag parity for all standard features, making it a
drop-in replacement that adds monorepo superpowers.
