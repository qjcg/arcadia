# DNS Skill with Doggo

This skill uses the `doggo` DNS client for querying DNS records.

## Installation

First, install doggo by running:

```bash
./install.sh
```

## Usage

The skill accepts JSON input via stdin with the following fields:

- `domain` (required): The domain to query
- `record_type` (optional): The DNS record type (e.g., 'A', 'MX', 'TXT')
- `dns_server` (optional): Specific DNS server to use (e.g., '8.8.8.8')

The program outputs the DNS query result in JSON format from doggo.

### Example

```bash
echo '{"domain": "example.com", "record_type": "A"}' | go run main.go
```

## Files

- `main.go`: The Go wrapper program
- `go.mod`: Go module file
- `install.sh`: Script to install doggo
- `test.sh`: Test script
