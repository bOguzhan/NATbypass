// File: internal/discovery/nat_types.go

package discovery

// NATType represents different NAT types that can be detected
type NATType string

const (
	// NATUnknown represents an unknown or undetected NAT type
	NATUnknown NATType = "unknown"

	// NATFullCone represents a full cone NAT (least restrictive)
	NATFullCone NATType = "full-cone"

	// NATAddressRestrictedCone represents an address-restricted cone NAT
	NATAddressRestrictedCone NATType = "address-restricted-cone"

	// NATPortRestrictedCone represents a port-restricted cone NAT
	NATPortRestrictedCone NATType = "port-restricted-cone"

	// NATSymmetric represents a symmetric NAT (most restrictive)
	NATSymmetric NATType = "symmetric"
)
