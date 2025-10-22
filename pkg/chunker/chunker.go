package chunker

import (
	"fmt"
	"math/big"
	"net"
)

// ChunkCIDR splits the given CIDR into subnets with the requested targetPrefix length.
// Assumptions:
// - targetPrefix is the desired prefix length (e.g. 26 for /26)
// - targetPrefix must be >= original prefix and <= address bit length (32 or 128)
// Returns a slice of CIDR strings covering the original network.
func ChunkCIDR(cidr string, targetPrefix int) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("invalid cidr: %w", err)
	}

	var maxBits int
	if ip.To4() != nil {
		maxBits = 32
	} else {
		maxBits = 128
	}

	origPrefix, _ := ipnet.Mask.Size()
	if targetPrefix < origPrefix {
		return nil, fmt.Errorf("target prefix %d is smaller than original prefix %d", targetPrefix, origPrefix)
	}
	if targetPrefix > maxBits {
		return nil, fmt.Errorf("target prefix %d is larger than max bits %d", targetPrefix, maxBits)
	}
	if targetPrefix == origPrefix {
		return []string{ipnet.String()}, nil
	}

	diff := targetPrefix - origPrefix
	// sanity check to avoid insane allocations (protect against huge splits)
	if diff > 32 {
		return nil, fmt.Errorf("requested split too large: diff=%d", diff)
	}

	// count = 1 << diff
	count := new(big.Int).Lsh(big.NewInt(1), uint(diff))
	// step = 1 << (maxBits - targetPrefix)  (addresses per subnet)
	step := new(big.Int).Lsh(big.NewInt(1), uint(maxBits-targetPrefix))

	base := ipToBigInt(ip)

	one := big.NewInt(1)
	res := make([]string, 0)
	for i := big.NewInt(0); i.Cmp(count) < 0; i.Add(i, one) {
		offset := new(big.Int).Mul(i, step)
		cur := new(big.Int).Add(base, offset)
		curIP := bigIntToIP(cur, maxBits)
		ipnet := &net.IPNet{IP: curIP, Mask: net.CIDRMask(targetPrefix, maxBits)}
		res = append(res, ipnet.String())
	}

	return res, nil
}

// ChunkCIDRBySize splits the given CIDR into subnets each containing `size` IP addresses.
// `size` must be a positive power of two (e.g., 1,2,4,8,16...) and fit within the address space.
// It computes the corresponding prefix length and delegates to ChunkCIDR.
func ChunkCIDRBySize(cidr string, size int) ([]string, error) {
	if size <= 0 {
		return nil, fmt.Errorf("size must be > 0")
	}
	// size must be power of two
	if size&(size-1) != 0 {
		return nil, fmt.Errorf("size must be a power of two")
	}

	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("invalid cidr: %w", err)
	}
	var maxBits int
	if ip.To4() != nil {
		maxBits = 32
	} else {
		maxBits = 128
	}

	// compute prefix: size = 1 << (maxBits - prefix) => prefix = maxBits - log2(size)
	// compute log2(size)
	log := 0
	v := size
	for v > 1 {
		v >>= 1
		log++
	}
	targetPrefix := maxBits - log

	// validate that targetPrefix is not smaller than original prefix
	origPrefix, _ := ipnet.Mask.Size()
	if targetPrefix < origPrefix {
		return nil, fmt.Errorf("requested size yields prefix /%d which is smaller than original prefix /%d", targetPrefix, origPrefix)
	}

	return ChunkCIDR(cidr, targetPrefix)
}

// ipToBigInt converts net.IP to big.Int (treats IPv4 as 4 bytes, IPv6 as 16 bytes)
func ipToBigInt(ip net.IP) *big.Int {
	// Prefer IPv4 4-byte representation when possible to avoid embedded-mapping issues.
	if v4 := ip.To4(); v4 != nil {
		return new(big.Int).SetBytes(v4)
	}
	b := ip.To16()
	if b == nil {
		// Shouldn't happen, but fallback to whatever raw bytes exist.
		b = ip
	}
	return new(big.Int).SetBytes(b)
}

// bigIntToIP converts a big.Int back to net.IP, producing 4-byte IPv4 or 16-byte IPv6 based on maxBits
func bigIntToIP(i *big.Int, maxBits int) net.IP {
	byteLen := 16
	if maxBits == 32 {
		byteLen = 4
	}
	b := i.Bytes()
	// left-pad with zeros to required length
	if len(b) < byteLen {
		padded := make([]byte, byteLen)
		copy(padded[byteLen-len(b):], b)
		b = padded
	} else if len(b) > byteLen {
		// Keep the least-significant bytes (right-most) when truncating.
		b = b[len(b)-byteLen:]
	}
	if byteLen == 4 {
		return net.IPv4(b[0], b[1], b[2], b[3])
	}
	// For IPv6 return a 16-byte IP value
	return net.IP(b)
}
