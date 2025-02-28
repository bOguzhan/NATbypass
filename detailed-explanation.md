
---

### **`detailed-project.md`** (Detailed Project Design)

```markdown
# Peer-to-Peer NAT Traversal System Design

## Overview

This project implements a **Peer-to-Peer NAT Traversal** system aimed at enabling direct communication between clients after discovering their public IPs and ports through an intermediary **Mediatory Server**. The goal is to minimize server egress, leverage **hole-punching** for NAT traversal, and eliminate the need for a fallback **TURN server** if NAT traversal fails.

## High-Level Architecture

1. **Mediatory Server**:
   - The Mediatory Server is responsible for initiating the communication process by providing clients with their public-facing IP and port.
   - The Mediatory Server is a **signaling server** that communicates with clients via **TCP** (HTTP or WebSocket).

2. **Application Server**:
   - After the Mediatory Server facilitates the handshake, the **Application Server** facilitates peer-to-peer communication by coordinating NAT hole-punching.
   - The **Application Server** handles data exchange between clients using **UDP** once the NAT traversal is successful.

3. **Client Communication**:
   - The client connects to the **Mediatory Server** to exchange information, receives the public-facing IP/port of the **Application Server**, and then directly communicates with another client using **UDP**.

4. **Key Goals**:
   - **Minimal Egress Traffic**: The server is only used for the initial signaling and public IP discovery.
   - **No TURN Server**: If NAT traversal fails, there is no fallback. The connection attempt is simply abandoned.
   - **Efficient, Scalable System**: Designed using **Go** for fast, efficient concurrency and **GCP** for cloud infrastructure scalability.

## Step-by-Step System Flow

### 1. **Client Initiates Communication**

- The client sends a request to the **Mediatory Server** with its private address and port.
- **Mediatory Server** receives the request and discovers its public IP/port using **STUN** (Session Traversal Utilities for NAT).

### 2. **Mediatory Server Responds with Public IP**

- The **Mediatory Server** uses **STUN** to discover its own public IP and port and responds to the client with this information.
- The client now knows the public IP and port of the server.

### 3. **Client Sends an Outbound Packet**

- The client sends an outbound packet to the **Application Server** using the public IP and port obtained from the **Mediatory Server**.
- This establishes a NAT mapping on the client’s NAT device for the **Application Server's** IP and port.

### 4. **Server Sends Back an Outbound Packet to Client**

- The **Application Server** sends an outbound packet to the client’s public IP and port. If the client’s NAT device permits incoming traffic from the server’s IP/port (since the client initiated the communication), the NAT hole-punching process succeeds.
  
### 5. **Hole Punching Successful – Direct Communication**

- Once both **client** and **server** send packets to each other, their NAT devices will allow communication. 
- Peer-to-peer communication happens directly using **UDP** or **WebRTC** (for higher-level abstraction).

### 6. **Failure Scenario**

- If the **hole-punching** process fails (e.g., due to restrictive NAT devices), the connection attempt is abandoned.
- **No fallback** to a **TURN server** is available. The client detects failure via timeout or lack of incoming traffic.

## System Components

### **Mediatory Server**:

- **Role**: Facilitates signaling and NAT discovery between client and server.
- **Technology**: 
  - **Go (Gin Framework)** for the HTTP signaling layer.
  - **STUN (pion/stun)** for discovering public IP and port.
  - **GCP** (Compute Engine or App Engine) for hosting.

### **Application Server**:

- **Role**: Handles peer-to-peer communication after initial handshake.
- **Technology**:
  - **Go** for network communication (using **UDP** for peer-to-peer).
  - **STUN (pion/stun)** for NAT traversal.
  - **WebRTC** (optional) for additional peer-to-peer capabilities.
  
### **Database** (Optional):

- **Firestore/Datastore** (for metadata storage, e.g., client information, NAT types).

## Security Considerations

- **TLS/SSL**: All communication between the client and server will be encrypted using **HTTPS**.
- **Authentication**: Use **JWT** for securing client access.
- **Firewall & Network**: Configure GCP firewall rules to restrict external access to the Mediatory Server.

## Deployment Strategy

- **Cloud Platform**: Use **Google Cloud Platform (GCP)** for hosting the **Mediatory Server**.
  - Consider using **Google Kubernetes Engine (GKE)** or **App Engine** for scaling the Mediatory Server.
  - **Docker** for containerization, ensuring consistent deployment.

- **CI/CD Pipeline**: Set up **GitHub Actions** or **GitLab CI** for automated testing and deployment.
  - Docker-based CI pipelines to automate the building and deployment process.

## Monitoring & Logging

- Use **Google Cloud Monitoring** to monitor server health and errors.
- Implement **Prometheus/Grafana** for custom metrics and visualization.

## Conclusion

This system design ensures direct communication between clients with minimal server egress, optimized NAT traversal using hole-punching techniques, and a simple architecture that is easy to maintain and scale. By leveraging **Go** and **GCP**, the system is both efficient and secure.

