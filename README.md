# Protocol Visualizer

This repository contains the source code for the 2026 BSc Project "Protocol Visualisation" in Software Development at the IT University of Copenhagen.

# Live demo

A hosted version of the app is available at https://protovis.villeret.dk/

# Running locally

The app can be run either directly with Bun or via Docker.

## Using Docker

```sh
docker run -p 3000:3000 ghcr.io/jonathanvil/protocol-visualizer
```

The app will be available at `http://localhost:3000`.

## Using Bun

### Prerequisites

- [Bun](https://bun.sh) — used as the package manager and runtime
- [Node.js](https://nodejs.org) — required by the SvelteKit adapter at runtime

```sh
bun install
bun run dev
```

The app will be available at `http://localhost:5173`.
