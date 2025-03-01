DISCLAIMER. THIS CODE IN THIS PROJECT IS %100 WRITTEN BY CLAUDE 3.7 ON COPILOT ON VSCODE. (this sentence is written by me)
# NATbypass: Multi-Protocol P2P NAT Traversal System

A peer-to-peer NAT traversal system that enables direct client-to-client communication through UDP and TCP protocols with minimal server involvement.

## Overview

NATbypass enables direct peer-to-peer connections between clients behind NAT without relying on relay servers, supporting both UDP and TCP protocols. The system intelligently selects the optimal protocol based on NAT configuration and provides fallback mechanisms when needed.

### Key Features

- **Multi-Protocol Support**: Works with both UDP and TCP protocols
- **Protocol-Agnostic API**: Single unified API with protocol-specific implementations
- **Smart Protocol Selection**: Automatic selection based on NAT type and network conditions
- **Fallback Mechanisms**: Gracefully fall back to alternative protocol when primary fails
- **Minimal Server Dependency**: Direct P2P connections with only signaling through server
- **No TURN Required**: Eliminates need for bandwidth-intensive TURN relay servers
- **NAT Type Detection**: Identifies client NAT configuration for optimal strategy selection

## Architecture

The system consists of two main components:

1. **Mediatory Server**
   - Provides signaling service for clients to discover each other
   - Implements STUN functionality for NAT type detection
   - Facilitates initial connection negotiation and protocol selection

2. **Application Server**
   - Enables NAT hole-punching for both UDP and TCP
   - Manages connection tracking and state
   - Facilitates direct P2P communication

## Protocol Support

### UDP Traversal
- Traditional UDP hole punching
- Optimized for connectionless communication
- Lower latency, higher success rate

### TCP Traversal
- TCP simultaneous open technique
- Connection-oriented with reliable delivery
- Better compatibility with restrictive firewalls

## Development

### Prerequisites
- Go 1.18+
- Docker and Docker Compose
- Git
- Access to Google Cloud Platform (for deployment)

### Local Development

1. Clone the repository
```bash
git clone https://github.com/yourusername/NATbypass.git
cd NATbypass
```

Install dependencies
Run locally
Run with Docker
Protocol Selection
The system automatically selects the optimal protocol based on:

NAT type detection
Network conditions
Application requirements
You can also explicitly select your preferred protocol through configuration.

Deployment
Google Cloud Platform
Set up GCP project
Deploy services
License
MIT License

Contributing
Contributions are welcome! Please feel free to submit a Pull Request.