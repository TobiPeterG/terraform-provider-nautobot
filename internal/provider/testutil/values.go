package testutil

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"net"
	"testing"
	"time"
)

func AccSeedForTest(t *testing.T) int64 {
	hash := fnv.New64a()
	_, _ = hash.Write([]byte(t.Name()))
	nameHash := hash.Sum64()

	return int64(uint64(time.Now().UnixNano()) ^ nameHash)
}

func AccPrefixCIDR(seed int64, offset int) string {
	hash := fnv.New64a()

	var input [16]byte
	binary.LittleEndian.PutUint64(input[0:8], uint64(seed))
	binary.LittleEndian.PutUint64(input[8:16], uint64(offset))
	_, _ = hash.Write(input[:])

	value := hash.Sum64()
	secondOctet := int(value & 0xFF)
	thirdOctet := int((value >> 8) & 0xFF)
	fourthOctet := int((value >> 16) & 0xF0)
	if secondOctet == 0 {
		secondOctet = 1
	}
	if thirdOctet == 0 {
		thirdOctet = 1
	}

	return fmt.Sprintf("21.%d.%d.%d/28", secondOctet, thirdOctet, fourthOctet)
}

func AccIPv6PrefixCIDR(seed int64, offset int) string {
	hash := fnv.New64a()

	var input [16]byte
	binary.LittleEndian.PutUint64(input[0:8], uint64(seed))
	binary.LittleEndian.PutUint64(input[8:16], uint64(offset))
	_, _ = hash.Write(input[:])

	value := hash.Sum64()
	return fmt.Sprintf("2001:db8:%x:%x::/124", uint16(value), uint16(value>>16))
}

func AccIPRangeBounds(cidr string) (string, string) {
	ip, _, _ := net.ParseCIDR(cidr)
	ip = ip.To4()
	return ipv4AddressAtOffset(ip, 0), ipv4AddressAtOffset(ip, 2)
}

// AccAvailableIPRangeBounds returns a small range containing usable host
// addresses. AccIPRangeBounds intentionally includes the network address and
// is useful for testing IP address ranges themselves, but Nautobot will not
// allocate that address from a prefix.
func AccAvailableIPRangeBounds(cidr string) (string, string) {
	ip, _, _ := net.ParseCIDR(cidr)
	ip = ip.To4()
	return ipv4AddressAtOffset(ip, 1), ipv4AddressAtOffset(ip, 3)
}

func ipv4AddressAtOffset(networkIP net.IP, offset int) string {
	return fmt.Sprintf("%d.%d.%d.%d", networkIP[0], networkIP[1], networkIP[2], int(networkIP[3])+offset)
}

func AccVLANVID(seed int64, offset int) int {
	base := int(uint64(seed) % 2000)
	return 1000 + base + offset
}
