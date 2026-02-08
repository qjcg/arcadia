# sv

**One repo, many versions, zero friction.**

`sv` is a semantic versioning tool built for monorepos. It extends the
core philosophy of `svu` by introducing path-based tagging, allowing
you to manage independent versions for multiple modules within a
single repository.

## Key Features

- **Path-Scoped Discovery**: Automatically detects module boundaries and filters tags by path (e.g., `x/slidesdeck/v2.3.1`).
- **Conventional Commits**: Uses your commit history to determine the next major, minor, or patch version.
- **Monorepo Ready**: Designed to work from the root of a large repository or deep within a module subdirectory.
- **`svu` Compatible**: Familiar CLI patterns for a seamless transition.

## Multi-Module Repositories

In a Go monorepo (a repository containing multiple `go.mod` files), each module must be tagged and versioned independently. To associate a tag with a specific module, Go uses path-based tagging (e.g., `x/slidesdeck/v1.2.3`). `sv` automates this process by detecting module boundaries and applying the correct path prefix to tags.

For more information, see the [official Go Wiki on Multi-Module Repositories](https://go.dev/wiki/Modules#faqs--multi-module-repositories).

## Quickstart

### View current version for a module
```bash
# Inside a module directory
sv current
```

### Calculate next version
```bash
sv next
```

### Recursive discovery
```bash
# From the repository root
sv next --all
```

## Documentation

Full documentation is available in the [docs](./docs) directory:
- [Elevator Pitch](./docs/PITCH.md)
- [PRD](./docs/PRD.md)
- [SRS](./docs/SRS.md)
- [PR-FAQ](./docs/PR-FAQ.md)

## Automation

### GitHub Actions
To automate tagging on merge to `main`, use `sv next --all` and loop through the output:

```yaml
- name: Auto Tag Modules
  run: |
    sv next --all | while read -r line; do
      tag=$(echo "$line" | awk '{print $3}')
      git tag "$tag"
    done
    git push --tags
```

### Lefthook
You can use `sv` with `lefthook` to alert developers of pending version bumps before pushing:

```yaml
# lefthook.yml
pre-push:
  commands:
    sv-check:
      run: sv next --all
```
