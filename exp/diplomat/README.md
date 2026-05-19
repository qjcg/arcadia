# Diplomat

Generate PDF diplomas and certificates of completion.

## Features

- Generates professional PDF certificates in letter landscape format (11" × 8.5")
- Customizable borders with navy outer frame and gold inner accent
- Embedded Droid Sans font for consistent rendering
- Supports multiple recipients in a single batch
- Unicode support for international names

## Installation

```bash
go build -o bin/diplomat .
```

## Usage

### Generate diplomas

```bash
diplomat generate \
    -c "Fun with JavaScript" \
    -i "Jane Instructor" \
    -p "July 12-15 2015 (22.5 hours)" \
    -r "Alice Smith,Bob Jones,Jean Étudiant"
```

### Options

| Flag           | Short | Description                     | Default             |
|----------------|-------|---------------------------------|---------------------|
| `--course`     | `-c`  | Course name                     | required            |
| `--instructor` | `-i`  | Instructor name                 | "Rory Q. Teachalot" |
| `--period`     | `-p`  | Training period/dates           | required            |
| `--recipients` | `-r`  | Comma-separated recipient names | "Joe Learnery"      |
| `--output-dir` | `-o`  | Output directory                | "diplomas"          |
| `--dry-run`    |       | Preview without generating PDFs | false               |
| `--json`       |       | Output results as JSON          | false               |
| `--quiet`      | `-q`  | Suppress non-error output       | false               |
| `--verbose`    | `-v`  | Show detailed progress          | false               |

### Output

Generates one PDF per recipient in `<output-dir>/<slugified-course-name>/`:

```
diplomas/
└── fun-with-javascript/
    ├── alice-smith.pdf
    ├── bob-jones.pdf
    └── jean-etudiant.pdf
```

### JSON Output

Use `--json` to output the diploma configuration:

```bash
diplomat generate -c "Go Basics" -i "Bill G." -p "Jan 2024" -r "Alice" --json
```

```json
{
  "Course": "Go Basics",
  "Period": "Jan 2024",
  "Instructor": "Bill G.",
  "Recipients": [
    "Alice"
  ],
  "OutputDir": "./diplomas/go-basics"
}
```

## Design

Diplomas are generated with:

- **Page size**: Letter landscape (792 × 612 points)
- **Background**: Cream/off-white (#FFFEF5)
- **Navy outer border**: 4pt frame with 8pt margins
- **Gold inner border**: 1.5pt accent with 14pt margins
- **Typography**: Droid Sans with navy (#1A365D) headings and gold (#C9A227) accents

## Configuration

The `DiplomaSet` struct drives generation:

```go
type DiplomaSet struct {
    Session   // Course, Period, Instructor, Recipients
    OutputDir string
}
```
