# Contributing to Pam

We welcome contributions! Here's how to get started:

## Development Setup

You need Go 1.21+ and the database client libraries for the databases you want to test against.

### Option 1: Using Nix

Enter the development shell with all required dependencies:

```bash
nix-shell
```

This provides:
- Go compiler
- Database clients: PostgreSQL, SQLite, Oracle Instant Client, duckDB

#### Option 2: Manual Setup

Install Go and the database clients you need:

```bash
# Go
go install github.com/eduardofuncao/pam/cmd/pam@latest

# PostgreSQL client (Ubuntu/Debian)
sudo apt install postgresql-client

# SQLite
sudo apt install sqlite3

# Oracle Instant Client
# Download from https://www.oracle.com/database/technologies/instant-client/downloads.html
```

### Testing with Real Databases

To test Pam with actual database connections, use the
**[dbeesly](https://github.com/eduardofuncao/dbeesly)** project - a collection
of pre-configured database setups with sample data.


#### Example 

**SQL Server:**
```bash
git clone https://github.com/dbeesly/sqlserver.git
cd sqlserver
make start
pam init sqlserver-dev sqlserver "sqlserver://sa:MyStrongPass123@localhost:1433/master"
```

Each dbeesly database includes:

- A make file with start, stop, logs and clean commands (run `make start` to
start the db server), which run docker-compose under the hood if needed
- Pre-populated data (including the tables employees, departments, timesheets)

### Making Changes

1. Fork the repository
2. Create a feature branch: `git checkout -b my-feature`
3. Make your changes and test with dbeesly databases
4. Commit: `git commit -m "Add my feature"`
5. Push: `git push origin my-feature`
6. Open a Pull Request

### Development Workflow

```bash
# Build locally
go build -o pam ./cmd/pam

# Run your build
./pam status

# Test with a dbeesly database
cd ../dbeesly/sqlserver
make start
cd ../pam
./pam run "SELECT * FROM employees"
```
