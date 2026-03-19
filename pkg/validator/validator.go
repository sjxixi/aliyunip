package validator

import (
	"fmt"
	"net"
	"strings"
)

func ValidateIPv4(ip string) error {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return fmt.Errorf("invalid IPv4 address format: %s", ip)
	}

	for _, part := range parts {
		if len(part) == 0 || len(part) > 3 {
			return fmt.Errorf("invalid IPv4 address format: %s", ip)
		}

		if part[0] == '0' && len(part) > 1 {
			return fmt.Errorf("invalid IPv4 address format (leading zeros): %s", ip)
		}

		var octet int
		for _, c := range part {
			if c < '0' || c > '9' {
				return fmt.Errorf("invalid IPv4 address format: %s", ip)
			}
			octet = octet*10 + int(c-'0')
		}

		if octet < 0 || octet > 255 {
			return fmt.Errorf("invalid IPv4 address octet: %s", ip)
		}
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil || parsedIP.To4() == nil {
		return fmt.Errorf("invalid IPv4 address: %s", ip)
	}

	return nil
}

func ValidateCIDR(cidr string) error {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("invalid CIDR format: %s", cidr)
	}

	if ip.To4() == nil {
		return fmt.Errorf("CIDR must be IPv4: %s", cidr)
	}

	mask := ipnet.Mask
	ones, bits := mask.Size()
	if bits != 32 {
		return fmt.Errorf("CIDR must be IPv4: %s", cidr)
	}
	if ones < 0 || ones > 32 {
		return fmt.Errorf("invalid prefix length: %s", cidr)
	}

	return nil
}

func IsValidIPv4(ip string) bool {
	return ValidateIPv4(ip) == nil
}

func IsValidCIDR(cidr string) bool {
	return ValidateCIDR(cidr) == nil
}
