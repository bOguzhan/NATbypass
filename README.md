# Peer-to-Peer NAT Traversal System

This project implements a **Peer-to-Peer NAT Traversal** system with minimal server involvement. The system allows clients to directly communicate with each other after discovering their public IP and port via a **Mediatory Server**. The primary goal is to minimize server egress and use **hole-punching** techniques for establishing direct peer-to-peer communication.

## Architecture Overview

1. **Mediatory Server (Signaling Server)**:
   - Facilitates initial connection between client and server.
   - Provides clients with the public-facing IP address and port.
   - Handles **STUN** requests to discover its public IP and port.
   
2. **Application Server**:
   - Once the Mediatory Server has provided the public IPs/ports, clients will attempt direct communication.
   - **UDP** hole-punching is used for NAT traversal to establish direct communication.
   - Once the connection is established, no further server involvement is required.

3. **Tech Stack**:
   - **Programming Language**: Go (Golang)
   - **Web Framework**: Gin (for HTTP signaling)
   - **Cloud Provider**: Google Cloud Platform (GCP)
   - **Containerization**: Docker
   - **Networking**: UDP for direct peer-to-peer communication, STUN for NAT discovery

## Features

- **STUN-based NAT Discovery**: Clients use the server to discover public IPs and ports for NAT hole-punching.
- **Minimal Server Egress**: Server only facilitates signaling, no data relaying.
- **Hole-Punching**: Clients attempt to establish direct peer-to-peer communication using NAT hole-punching.
- **No Fallback to TURN Server**: If NAT traversal fails, the connection attempt is aborted.
- **Lightweight and Scalable**: Built using Go, a fast and efficient language for handling concurrent requests.

## Requirements

- **Go**: v1.18 or later
- **Docker**: For containerization and deployment
- **Google Cloud Platform**: For hosting the Mediatory Server (Compute Engine, App Engine, or Kubernetes Engine)

## Getting Started

### 1. Clone the repository:

```bash
git clone https://github.com/yourusername/peer-to-peer-nat-traversal.git
cd peer-to-peer-nat-traversal
```
2. Set up the Mediatory Server (Signaling Server)

    Use GCP's Compute Engine or App Engine to deploy the Mediatory Server.
    The Mediatory Server handles STUN requests and returns public-facing IP and port information.

3. Build the Application Server

    The Application Server uses Go for peer-to-peer communication.
    After the initial handshake, clients directly communicate via UDP.

4. Run the Application

    Build the Go server:

go build -o mediatory-server ./cmd/mediatory-server
go build -o application-server ./cmd/application-server

    Run the Mediatory Server:

./mediatory-server

    Run the Application Server:

./application-server

5. Docker Setup (Optional)

To deploy using Docker for both the Mediatory Server and Application Server, create a Docker image for each component and deploy them:

docker build -t mediatory-server ./cmd/mediatory-server
docker build -t application-server ./cmd/application-server

Then, run the containers:

docker run -d -p 8080:8080 mediatory-server
docker run -d -p 8081:8081 application-server

Testing
Unit Tests:

go test ./...

Integration Tests:

Ensure that both Mediatory Server and Application Server are able to establish direct communication between peers after the initial handshake.
Contributing

We welcome contributions! Please fork the repository and submit a pull request. For any bug reports or feature requests, create an issue.
License

This project is licensed under the MIT License - see the LICENSE file for details.