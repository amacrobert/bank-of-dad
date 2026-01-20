# Bank of Dad

Teach your kids the power of compounding interest with Bank of Dad!

## Prerequisites

- Docker
- Docker Compose

## Project Structure

```
bank-of-dad/
├── backend/          # Go HTTP server
├── frontend/         # React + Vite + TypeScript
└── docker-compose.yaml
```

## Local Development

### Running with Docker Compose (Hot Reloading)

Build and start all services with hot reloading enabled:

```bash
docker compose up --build
```

This automatically uses `docker-compose.override.yaml` which enables:
- **Frontend**: Vite dev server with HMR (Hot Module Replacement) - changes to React components update instantly in the browser

Access the application:
- Frontend: http://localhost:8000
- Backend API: http://localhost:8001

### Development Workflow

1. Start the services: `docker compose up --build`
2. Edit frontend code in `frontend/src/` - browser updates automatically
3. View logs in the terminal for build/reload status

### Running Production Build

To run the production build (without hot reloading):

```bash
docker compose -f docker-compose.yaml up --build
```

### Stopping Services

```bash
docker compose down
```

## API Endpoints

### GET /message

Returns a JSON message.

**Response:**
```json
{
  "message": "Hello from Bank of Dad!"
}
```
