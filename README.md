# eng_club

Go backend for the English club: sessions, pairings, MySQL persistence, and an HTTP API (`cmd/api_server`).

## Database migrations

Schema changes live under [`database/migrations/`](database/migrations/). They are **not** applied when the app opens the database; run the migrate command after setup or whenever new SQL files are added.

From the repository root (with `config.json` in the working directory, or copy from [`config.example.json`](config.example.json)):

```bash
go run ./cmd/migrate
```

Optional path to a config file:

```bash
go run ./cmd/migrate /path/to/config.json
```

Then start the API server (or other commands) as usual.

## API server

```bash
go run ./cmd/api_server
```

Uses the same `config.json` for the database and HTTP `server.host` / `server.port`.

## Admin frontend (Vue 3 + Bootstrap 5)

Located in [`frontend/`](frontend/). You need Node.js and npm installed locally. Vite proxies `/api` to `http://127.0.0.1:8080` (change in [`frontend/vite.config.js`](frontend/vite.config.js) if needed).

```bash
cd frontend
npm install   # installs frontend/node_modules from package.json
npm run dev
```

Then open the printed local URL (e.g. `http://localhost:5173`). Run the API server on port `8080` in another terminal.

Production build:

```bash
cd frontend
npm run build
```
