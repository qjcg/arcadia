# fin

**Versatile financial calculator** — tax, interest, and more.

`fin` is a CLI for financial computations. Currently ships with Canadian
income tax penalty and interest calculators (CRA & Revenu Québec), with
more tools planned.

## Features

- **Tax penalties & interest**: Late-filing penalties and prescribed interest for CRA and Revenu Québec, computed with daily compounding using official quarterly rates
- **Embedded rates**: Bundled rate database kept current through built-in fetch/update commands
- **JSON output**: Machine-readable output for scripting and automation

## Installation

```bash
go build -o fin .
```

## Usage

### Calculate tax penalties and interest

```bash
# $5000 due to CRA, filed 1.5 months late, paid 2 months late
fin tax pi \
    --year 2024 \
    --base-due-cra 5000 \
    --expected-filing-date 2024-04-30 \
    --actual-filing-date 2024-06-15 \
    --expected-payment-date 2024-04-30 \
    --actual-payment-date 2024-06-15

# JSON output
fin tax pi --year 2024 --base-due-cra 5000 --output json
```

### View interest rates

```bash
fin tax pi rates
fin tax pi rates --jurisdiction cra
fin tax pi rates --year 2024 --jurisdiction rq
```

### Update rates database

```bash
fin tax pi update --range 2020-2025 --source cra
fin tax pi update --range 2020-2025 --source both --dry-run
```

## Current Modules

```
fin
├── tax
│   └── penalties-and-interest (pi)
│       ├── rates       — display embedded prescribed interest rates
│       └── update      — fetch latest rates from official sources
└── (more coming)
```

## Design

- **Money**: Built on `shopspring/decimal` for exact financial arithmetic
- **Rules**: Per-jurisdiction penalty formulas with daily compounding interest
- **Rates DB**: Embedded YAML database at `internal/rates/rates.yaml`, versioned and updatable via CLI

## Rates Database

Interest rates are embedded at build time. Use `fin tax pi update` to fetch
the latest prescribed rates from CRA and Revenu Québec official sources.