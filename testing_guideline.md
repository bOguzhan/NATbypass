NAT Bypass Testing Guidelines
Overview
This document provides guidelines for testing the NAT Bypass system across different protocols and network configurations. Based on our development plan, we're currently in Phase 5 (Testing and Validation), with several components already implemented and ready for testing.

Test Categories
1. Unit Testing
Run individual component tests using Go's testing framework:

Key test areas:

NAT discovery mechanisms
STUN client integration
Signaling API endpoints
Connection registry operations
Protocol-specific packet handling
2. Local Integration Testing
For quick local development testing:

This tests NAT traversal strategies in a simulated local environment without actual network constraints.

3. Protocol-Specific Testing
UDP Testing
Test UDP hole punching and related functionality:

Focus areas:

UDP hole punching across different NAT types
Packet loss resilience
Connection establishment timing
TCP Testing
Test TCP traversal techniques:

Focus areas:

TCP simultaneous open
Connection state management
Fallback mechanisms
4. Real-World NAT Scenario Testing
Test with Docker-simulated NAT environments:

This creates multiple Docker containers with different NAT configurations to simulate real-world conditions.

5. NAT Strategy Comparison Testing
Compare effectiveness of different traversal strategies:

Focus on evaluating:

Success rates per strategy
Connection establishment time
Stability of established connections
Docker Test Environment
For consistent testing across NAT types:

This creates a test environment with:

Full-cone NAT
Restricted-cone NAT
Port-restricted NAT
Symmetric NAT
Test Matrix
NAT Type A	NAT Type B	UDP Success Rate	TCP Success Rate
Full-cone	Full-cone	Expected: 95%+	Expected: 90%+
Full-cone	Restricted	Expected: 90%+	Expected: 85%+
Restricted	Restricted	Expected: 85%+	Expected: 80%+
Symmetric	Any	Expected: <50%	Expected: <40%
Reporting Issues
When reporting test failures:

Specify the test script that failed
Note the NAT configuration being tested
Include relevant logs from both server and client
Note the protocol being tested (UDP/TCP)
Comprehensive Testing
To run all tests:

This will execute all test scripts in sequence and generate a consolidated report.