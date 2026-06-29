I want to add a changelog feature to the sv tool.


# Features

- Use the `sv changelog` subcommand
  - `sv changelog` - prints full changelog to stdout
  - `sv changelog --from v0.2.0 --to v0.6.0` - prints changelog
  - `sv changelog --since 2025` - prints changelog since 2025 (starting on Jan 1 if not specified)
  - `sv changelog --since 8w` - prints changelog since 8 weeks ago
  - `sv changelog -d foo`  or `sv changelog --dir foo`
    - when this flag is provided, write changelog entries as files to the specified directory (see below). Default is do not write any files, output to stdout only
- generates changelogs in [keepachangelog.com format](https://keepachangelog.com/en/1.1.0/) from the repo's [conventional commits](https://www.conventionalcommits.org/en/v1.0.0/)
- stores changelog entries in a modular fashion and assembles them:
  - example entries:
    - v0.1.0.md - changelog for v0.1.0 in keepachangelog format
    - v0.1.0_overview.md - commentary on the v0.1.0 changes, the stuff before "Added", "Changed", "Fixed", etc. Can be written by human or AI or both.
    - v0.2.0.md
    - v0.2.0_overview.md
    - unreleased.md
    - unreleased_overview.md
  - assembled into:
    - CHANGELOG.md
- changes must be formatted as follows: `- $SHORT_COMMIT_HASH - $COMMIT_MESSAGE`
  - If the `-u` or `--url-prefix` flag is provided with an argument, it's used to turn the $SHORT_COMMIT_HASH into a markdown link by using the url prefix as a base
