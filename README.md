# Protocol Visualizer

This repository contains the source code for the 2026 BSc Project "Protocol Visualisation" in Software Development at the IT University of Copenhagen.

# Live demo

A hosted version of the app is available at https://protovis.villeret.dk/

# How it works

You implement the actors of a protocol in Go using the simulator in `go/simulator`, and run your
program. It serves the simulation over a WebSocket on port 8067 (see `PROTOCOL.md`), and this web
app connects to it to visualize and control the simulation.

`go/examples/pingpong` is a complete example.

# Running locally

## 1. Run a simulation

### Prerequisites

- [Go](https://go.dev) 1.26 or newer

```sh
cd go/examples/pingpong
go run .
```

## 2. Run the web app

The web app can be run either directly with Bun or via Docker.

### Using Bun

#### Prerequisites

- [Bun](https://bun.sh) — used as the package manager and runtime

```sh
cd web
bun install
bun run dev
```

The app will be available at `http://localhost:5173`.

### Using Docker

```sh
docker run -p 3000:3000 ghcr.io/jonathanvil/protocol-visualizer
```

The app will be available at `http://localhost:3000`.

The web app connects to the simulation at `ws://localhost:8067/ws`, so it has to run on the same
machine as your Go program.
