# Sentinel

Sentinel is a high-performance distributed system monitor implemented in Go. It provides real-time monitoring of CPU and RAM usage across multiple remote computers using gRPC for efficient client-side streaming.

## Features

- **Distributed Architecture:** A central server gathers reports from multiple distributed agents.
- **Real-time Dashboard:** A web-based dashboard to visualize metrics from all connected agents.
- **Efficient Streaming:** Uses gRPC for low-latency, high-throughput metrics reporting.
- **Single Binary:** Both server and agent functionalities are packed into a single CLI tool.
- **Flexible Configuration:** Managed with Cobra and Viper, supporting both CLI flags and YAML configuration.

## Project Structure

- `cmd/`: CLI command definitions (root, server, agent).
- `internal/`: Core logic including metrics collection, storage, and the dashboard server.
- `proto/`: gRPC service definitions and generated code.
- `web/`: Frontend assets for the dashboard.
- `Dockerfile.*` & `docker-compose.yml`: Containerization for easy deployment and demos.

## Getting Started

### Prerequisites

- [Go](https://go.dev/doc/install) 1.25 or later
- [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/) (for the demo)

### Quick Start: Docker Demo

You can quickly test Sentinel with a pre-configured 1-server, 5-agent distributed system.

1.  **Run the demo** with a single command from the project's root directory:
    ```bash
    docker-compose up --build
    ```
2.  **View the dashboard:** Once the containers are running, open your web browser and go to:
    [http://localhost:8080](http://localhost:8080)
3.  **Stop the demo:** To stop and remove all the containers and the network, open a new terminal and run:
    ```bash
    docker-compose down
    ```

## Manual Usage

### Building from source

```bash
make build
```

### Running the Server

Start the central server to collect metrics and host the dashboard:

```bash
./sentinel server --port 50051 --http-port 8080
```

### Running an Agent

Start an agent to report stats from a distant computer:

```bash
./sentinel agent --agent-id "node-01" --server-addr "localhost:50051" --interval 5
```

## Configuration

Sentinel can be configured using CLI flags or a `.sentinel.yaml` configuration file.

Example `.sentinel.yaml`:
```yaml
server:
  port: "50051"
  http-port: "8080"
agent:
  server-addr: "localhost:50051"
  agent-id: "my-agent"
  interval: 5
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
