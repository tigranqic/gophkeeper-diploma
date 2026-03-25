# GophKeeper

GophKeeper is a secure, end-to-end encrypted (E2EE) password manager and binary data storage system. It features a client-server architecture built with Go, gRPC, and PostgreSQL, designed to keep your secrets safe even if the server is compromised.

## Key Features

- **End-to-End Encryption (E2EE):** Data is encrypted on the client side using AES-256-GCM before being sent to the server. The server never sees your raw data.
- **Secure Authentication:** Uses `scrypt` for key derivation. Your master password never leaves your device.
- **Robust Synchronization:** Supports multi-device synchronization using server-side revisions and gRPC streaming for handling large datasets efficiently.
- **Zero-Knowledge Architecture:** The server stores only encrypted blobs (`bytes`). It has no knowledge of the content (whether it's a password, card, or binary file).
- **Memory Safety:** Sensitive data (passwords, keys) is actively zeroed out in memory immediately after use.

## Architecture

- **Protocol:** gRPC (Protobuf v3) with bidirectional streaming support.
- **Database:** PostgreSQL for persistent storage of users and encrypted records.
- **Encryption:**
  - **DEK (Data Encryption Key):** Derived from the Master Password using `scrypt`. Used to encrypt local data.
  - **Auth Hash:** A separate hash derived from the Master Password, used for server authentication (hashed again on the server with `bcrypt`).
- **Versioning:** Server-side `BIGSERIAL` revisions ensure strict ordering and conflict resolution ("Last Write Wins" based on revision).

## Prerequisites

- **Go** 1.25 or higher
- **PostgreSQL** 14+
- **Make** (optional, for convenience)

## Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/tigranqic/gophkeeper-diploma.git
cd gophkeeper-diploma
```

### 2. Build the Project

Use the provided `Makefile` to build both the server and the client:

```bash
make all
```

This will create `gophkeeper-server` and `gophkeeper-client` binaries in the `bin/` directory.

### 3. Database Setup

Ensure you have a PostgreSQL database running. You can start one quickly using Docker:

```bash
docker run --name gophkeeper-db -e POSTGRES_PASSWORD=postgres -p 5452:5432 -d postgres
```

Apply the migrations manually or using a tool like `migrate`. The initial schema is located in `migrations/000001_init.up.sql`.

For a quick start, you can run the SQL directly:

```bash
psql "postgres://postgres:postgres@localhost:5452/postgres?sslmode=disable" -f migrations/000001_init.up.sql
```

## Usage Guide

### Running the Server

Start the server, specifying the listening address and database DSN:

```bash
./bin/gophkeeper-server \
  -a localhost:8080 \
  -g localhost:3200 \
  -d "postgres://postgres:postgres@localhost:5452/postgres?sslmode=disable"
```

### Using the Client

The client is a CLI application. It stores encrypted data locally in `~/.gophkeeper.json`.

#### 1. Registration

Register a new user. You will be prompted to enter a **master password** — it is used both to derive the server auth hash and to encrypt your local data.

```bash
./bin/gophkeeper-client register -u myuser
```

> **Note:** The master password never leaves your device. Do **not** lose it — data recovery is impossible without it.

#### 2. Login

Login to download your data or sync changes.

```bash
./bin/gophkeeper-client login -u myuser
```

#### 3. Adding Data

Add a new login/password pair. You will be prompted for the master password if you haven't logged in recently.

```bash
./bin/gophkeeper-client add login --login "github.com" --pass "mysecretpass" --meta "Personal GitHub"
```

#### 4. Listing Records

View your local encrypted records (metadata is not shown encrypted here for clarity, but it is stored encrypted).

```bash
./bin/gophkeeper-client list
```

Output:
```text
ID: <uuid> | Type: RECORD_TYPE_LOGIN | Revision: 1 | Updated: 2023-10-27T10:00:00Z
```

#### 5. Decrypting Data

Retrieve and decrypt a specific record by its ID.

```bash
./bin/gophkeeper-client get <record-id>
```

#### 6. Synchronization

Sync your local changes with the server and download updates from other devices.

```bash
./bin/gophkeeper-client sync
```

## Security Model Details

1.  **Input Security:** The CLI uses `golang.org/x/term` to read passwords securely without echoing them to the terminal or saving them in shell history.
2.  **Memory Hygiene:** Critical variables (keys, passwords) are wiped from memory (`zeroed`) using a defer statement immediately after their scope ends.

## Development

### Running Tests

Run all unit tests with race detection:

```bash
make test
```

Check test coverage:

```bash
make test-coverage
```

### Protocol Buffers

If you modify `proto/gophkeeper.proto`, regenerate the Go code:

```bash
make proto
```
