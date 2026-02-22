# dude

Table of contents and single-section print for man pages.

## Usage

dude is just a simple wrapper for the man command, so just
use it with the arguments you would pass to man.

```sh
$ dude man
NAME
SYNOPSIS
DESCRIPTION
EXAMPLES
OVERVIEW
DEFAULTS
OPTIONS
   General options
   Main modes of operation
   Finding manual pages
   Controlling formatted output
   Getting help
EXIT STATUS
ENVIRONMENT
FILES
SEE ALSO
HISTORY

$ dude 7 man
NAME
SYNOPSIS
DESCRIPTION
   Title line
   Sections
   Fonts
   Other macros and strings
   Normal paragraphs
   Relative margin indent
   Indented paragraph macros
   Hypertext link macros
   Miscellaneous macros
   Predefined strings
   Safe subset
FILES
NOTES
BUGS
SEE ALSO
COLOPHON

# Print a section by name. Case insensitive.
$ dude bash files
FILES
       /bin/bash
              The bash executable
       /etc/profile
              The  systemwide  initialization  file, executed
              for login shells
       ~/.bash_profile
              The personal initialization file, executed  for
              login shells
       ~/.bashrc
              The  individual  per-interactive-shell  startup
              file
       ~/.bash_logout
              The individual login shell cleanup  file,  exe‐
              cuted when a login shell exits
       ~/.inputrc
              Individual readline initialization file

```
