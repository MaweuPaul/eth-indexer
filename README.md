# Ethereum Event Indexer

A real-time Ethereum event indexer that subscribes to ERC-20 logs, decodes
`Transfer` and `Approval` events, stores them in PostgreSQL, and broadcasts new
events to WebSocket clients.

The backend currently watches the Ethereum mainnet USDT contract
(`0xdAC17F958D2ee523a2206206994597C13D831ec7`). A Next.js frontend is included as
a starting point for a web dashboard.

## Features

- Live Ethereum log subscriptions through a WebSocket RPC provider
- ERC-20 `Transfer` and `Approval` event decoding
- PostgreSQL event persistence
- REST endpoint for querying events by contract
- WebSocket broadcast of newly indexed events
- Basic health check and HTML dashboard endpoint

## Project structure

```text
.
|-- Backend/               # Go indexer, API, and WebSocket server
|   |-- cmd/main.go        # Application entry point
|   |-- config/            # Environment configuration
|   `-- internal/          # API, decoder, hub, indexer, and store packages
`-- frontend/              # Next.js application
```

## Requirements

- Go 1.26.2 or a compatible version
- PostgreSQL
- An Ethereum WebSocket RPC URL, such as an Alchemy `wss://` endpoint
- Node.js and npm to run the optional frontend

## Backend setup

1. Create a PostgreSQL database and add the events table:

```sql
CREATE TABLE events (
    id BIGSERIAL PRIMARY KEY,
    contract TEXT NOT NULL,
    event_name TEXT NOT NULL,
    block_number BIGINT NOT NULL,
    tx_hash TEXT NOT NULL,
    log_index INTEGER NOT NULL,
    data JSONB NOT NULL,
    UNIQUE (tx_hash, log_index)
);

CREATE INDEX events_contract_block_idx
    ON events (contract, block_number DESC);
```

2. Create `Backend/.env`:

```dotenv
ALCHEMY_URL=wss://eth-mainnet.g.alchemy.com/v2/your-api-key
DATABASE_URL=postgres://postgres:password@localhost:5432/eth_indexer?sslmode=disable
PORT=8080
```

3. Start the service from the backend directory:

```bash
cd Backend
go mod download
go run ./cmd
```

The process connects to PostgreSQL, starts the HTTP server, and then listens for
new logs from the configured contract. The indexer processes events emitted
after it starts; it does not currently backfill historical blocks.

## API

With `PORT=8080`, the service exposes:

| Endpoint | Description |
| --- | --- |
| `GET /health` | Returns `{"status":"ok"}` |
| `GET /events/{contract}` | Returns stored events for a contract, newest block first |
| `GET /dashboard` | Serves the backend's basic HTML dashboard |
| `WS /ws` | Streams newly indexed events |

Example:

```bash
curl http://localhost:8080/events/0xdAC17F958D2ee523a2206206994597C13D831ec7
```

WebSocket messages use this shape:

```json
{
  "type": "new_event",
  "data": {
    "Contract": "0xdAC17F958D2ee523a2206206994597C13D831ec7",
    "EventName": "Transfer",
    "BlockNumber": 12345678,
    "TxHash": "0x...",
    "LogIndex": 1,
    "Data": {
      "event": "Transfer",
      "from": "0x...",
      "to": "0x...",
      "value": "1000000"
    }
  }
}
```

## Frontend

The `frontend` directory currently contains the default Next.js starter and is
not yet connected to the backend. To run it locally:

```bash
cd frontend
npm install
npm run dev
```

Open <http://localhost:3000>.

## Current limitations

- The watched contract is configured in `Backend/cmd/main.go` rather than via
  environment variables.
- Only standard ERC-20 `Transfer` and `Approval` events are decoded.
- There is no historical backfill or chain-reorganization handling yet.
- API routes do not currently implement authentication, pagination, or CORS.

## License

See [Backend/LICENSE](Backend/LICENSE).
