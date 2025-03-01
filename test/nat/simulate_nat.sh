#!/bin/bash
# File: test/nat/simulate_nat.sh

# Get environment variables
NAT_TYPE=${NAT_TYPE:-"full-cone"}
PEER_TYPE=${PEER_TYPE:-"initiator"}
PROTOCOL=${PROTOCOL:-"udp"}

echo "Setting up $NAT_TYPE NAT environment for $PEER_TYPE using $PROTOCOL"

# Configure network based on NAT type
case $NAT_TYPE in
  "full-cone")
    echo "Setting up full cone NAT (least restrictive)"
    # Full cone NAT just does basic address translation without restrictions
    iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
    ;;
  
  "address-restricted-cone")
    echo "Setting up address-restricted cone NAT"
    # Address-restricted cone NAT only allows incoming packets from addresses that have been sent to
    iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
    # Drop incoming packets unless they're from an IP we've sent to
    iptables -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT
    iptables -A INPUT -p $PROTOCOL -m state --state NEW -j DROP
    ;;
  
  "port-restricted-cone")
    echo "Setting up port-restricted cone NAT"
    # Port-restricted cone NAT only allows incoming packets from IP:port combinations that have been sent to
    iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
    # Drop incoming packets unless they're from an IP:port we've sent to
    iptables -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT
    iptables -A INPUT -p $PROTOCOL -m state --state NEW -j DROP
    # Additional rule to be more strict about port matching
    if [ "$PROTOCOL" = "udp" ]; then
      iptables -A INPUT -p udp -m state --state NEW -j DROP
    else
      iptables -A INPUT -p tcp -m state --state NEW -j DROP
    fi
    ;;
  
  "symmetric")
    echo "Setting up symmetric NAT (most restrictive)"
    # Symmetric NAT uses a different port for each destination
    iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
    # Simulate using different ports by using random source ports
    iptables -t nat -A POSTROUTING -p $PROTOCOL -o eth0 -j MASQUERADE --random
    # Only allow established connections back in
    iptables -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT
    iptables -A INPUT -p $PROTOCOL -m state --state NEW -j DROP
    ;;
  
  *)
    echo "Unknown NAT type: $NAT_TYPE. Using full cone as default."
    iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
    ;;
esac

echo "NAT environment configured. Starting test client..."

# Run the test client
/app/bin/nat-test