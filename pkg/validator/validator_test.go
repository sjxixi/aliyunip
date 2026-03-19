package validator

import (
	"testing"
)

func TestValidateIPv4_Valid(t *testing.T) {
	tests := []string{
		"0.0.0.0",
		"255.255.255.255",
		"192.168.1.1",
		"10.0.0.1",
		"172.16.0.1",
		"1.2.3.4",
		"127.0.0.1",
	}

	for _, ip := range tests {
		t.Run(ip, func(t *testing.T) {
			if err := ValidateIPv4(ip); err != nil {
				t.Errorf("ValidateIPv4(%q) unexpected error: %v", ip, err)
			}
			if !IsValidIPv4(ip) {
				t.Errorf("IsValidIPv4(%q) expected true, got false", ip)
			}
		})
	}
}

func TestValidateIPv4_Invalid(t *testing.T) {
	tests := []string{
		"",
		"256.0.0.1",
		"0.256.0.1",
		"0.0.256.1",
		"0.0.0.256",
		"192.168.1",
		"192.168.1.1.1",
		"192.168.1.a",
		"abc.def.ghi.jkl",
		"192..168.1.1",
		".192.168.1.1",
		"192.168.1.1.",
		"192.168.01.1",
		"192.168.1.01",
		"01.168.1.1",
		"192.016.1.1",
		"2001:db8::1",
		"::1",
	}

	for _, ip := range tests {
		t.Run(ip, func(t *testing.T) {
			if err := ValidateIPv4(ip); err == nil {
				t.Errorf("ValidateIPv4(%q) expected error, got nil", ip)
			}
			if IsValidIPv4(ip) {
				t.Errorf("IsValidIPv4(%q) expected false, got true", ip)
			}
		})
	}
}

func TestValidateCIDR_Valid(t *testing.T) {
	tests := []string{
		"0.0.0.0/0",
		"255.255.255.255/32",
		"192.168.1.0/24",
		"10.0.0.0/8",
		"172.16.0.0/12",
		"1.2.3.4/32",
		"127.0.0.0/8",
		"192.168.1.1/30",
	}

	for _, cidr := range tests {
		t.Run(cidr, func(t *testing.T) {
			if err := ValidateCIDR(cidr); err != nil {
				t.Errorf("ValidateCIDR(%q) unexpected error: %v", cidr, err)
			}
			if !IsValidCIDR(cidr) {
				t.Errorf("IsValidCIDR(%q) expected true, got false", cidr)
			}
		})
	}
}

func TestValidateCIDR_Invalid(t *testing.T) {
	tests := []string{
		"",
		"192.168.1.0",
		"192.168.1.0/",
		"192.168.1.0/33",
		"192.168.1.0/-1",
		"256.0.0.0/24",
		"192.256.1.0/24",
		"192.168.256.0/24",
		"192.168.1.256/24",
		"192.168.1/24",
		"192.168.1.1.1/24",
		"192.168.1.a/24",
		"abc.def.ghi.jkl/24",
		"192..168.1.0/24",
		".192.168.1.0/24",
		"192.168.1.0./24",
		"192.168.01.0/24",
		"2001:db8::1/64",
		"::1/128",
	}

	for _, cidr := range tests {
		t.Run(cidr, func(t *testing.T) {
			if err := ValidateCIDR(cidr); err == nil {
				t.Errorf("ValidateCIDR(%q) expected error, got nil", cidr)
			}
			if IsValidCIDR(cidr) {
				t.Errorf("IsValidCIDR(%q) expected false, got true", cidr)
			}
		})
	}
}
