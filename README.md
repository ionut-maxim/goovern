# Goovern

A TUI (Terminal User Interface) application for searching Romanian company data from ONRC (National Trade Register Office). Access company information via SSH and search through an interactive Bubble Tea interface.

## What it does

- **SSH-based TUI**: Connect via SSH to search Romanian company records
- **Full-text search**: Search companies by name or CUI (tax ID)
- **Background workers**: Automatically downloads and imports ONRC datasets from data.gov.ro
- **Structured data**: Stores company information in PostgreSQL for fast searching

## Running

### Prerequisites
- PostgreSQL database
- Go 1.23+ (if running without Docker)

### With Docker

```bash
# Build the Docker image
docker build -t goovern:latest .

# Run with PostgreSQL connection
docker run -d \
  --name goovern \
  -p 42069:42069 \
  -e GOO_DB_URL="postgresql://user:password@host:5432/dbname" \
  goovern:latest
```

### Without Docker

```bash
# Set environment variables
export GOO_DB_URL="postgresql://user:password@localhost:5432/dbname"

# Build and run
go build -o goovernd ./cmd/goovernd
./goovernd
```

### Connect

```bash
ssh localhost -p 42069
```

## Background Workers

Goovern uses [River](https://riverqueue.com/) for background job processing:

- **Update checker**: Periodically scans data.gov.ro for new ONRC datasets
- **Download worker**: Fetches CSV files from CKAN
- **Import worker**: Processes CSVs and loads data into PostgreSQL with dependency ordering

Workers respect data dependencies (e.g., `caen_versions` before `caen_codes`, `companies` before `company_status_history`).

## Security

The SSH server is built with [Wish](https://github.com/charmbracelet/wish) and accepts unauthenticated guest connections. Since all data is read-only public information from the National Trade Register, this configuration is secure for its intended use case.

## Configuration

Environment variables (prefix: `GOO_`):

- `GOO_DB_URL`: PostgreSQL connection string
- `GOO_LOG_LEVEL`: Log level (debug, info, warn, error)
- `GOO_LOG_TYPE`: Log format (pretty, json, text)

## License

MIT
