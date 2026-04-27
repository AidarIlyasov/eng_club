# Eng Club Admin (Vue)

- **Events** — list meetings, create (place + date + topic), open an event to add participants and **Start session** (POST `/api/events/:id/start`).
- **Places** — CRUD for venues (name, metro, map URL, image URL) used on the create-meeting screen.

## Prerequisites

[Node.js](https://nodejs.org/) and npm on your PATH (e.g. Node 22 + npm 10 is fine). You only need `npm install` **inside this folder** once (or after `package.json` changes) to install project dependencies into `frontend/node_modules`.

## Develop

Requires the Go API (`go run ./cmd/api_server` from repo root) so `/api` can be proxied.

```bash
cd frontend   # if not already here
npm install   # project deps, not installing Node itself
npm run dev
```

## Build

```bash
npm run build
```

Static output is written to `frontend/dist/` (serve behind your API or a reverse proxy as you prefer).
