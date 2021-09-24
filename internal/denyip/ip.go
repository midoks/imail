package denyip

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

type Checker struct {
	denyIPs    []*net.IP
	denyIPsNet []*net.IPNet
}

// NewChecker builds a new Checker given a list of CIDR-Strings to IPs.
func NewChecker(deniedIPs []string) (*Checker, error) {
	if len(deniedIPs) == 0 {
		return nil, errors.New("no IPs provided")
	}

	checker := &Checker{}

	for _, ipMask := range deniedIPs {
		if ipAddr := net.ParseIP(ipMask); ipAddr != nil {
			checker.denyIPs = append(checker.denyIPs, &ipAddr)
		} else {
			_, ipAddr, err := net.ParseCIDR(ipMask)
			if err != nil {
				return nil, fmt.Errorf("parsing CIDR IPs %s: %w", ipAddr, err)
			}
			checker.denyIPsNet = append(checker.denyIPsNet, ipAddr)
		}
	}

	return checker, nil
}

// IsAuthorized checks if provided request is authorized by the trusted IPs.
func (checker *Checker) IsAuthorized(addr string) error {
	var invalidMatches []string

	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}

	ok, err := checker.Contains(host)
	if err != nil {
		return err
	}

	if !ok {
		invalidMatches = append(invalidMatches, addr)
		return fmt.Errorf("%q matched none of the trusted IPs", strings.Join(invalidMatches, ", "))
	}
	return nil
}

// Contains checks if provided address is in the IPs.
func (checker *Checker) Contains(addr string) (bool, error) {
	if len(addr) == 0 {
		return false, errors.New("empty IP address")
	}

	ipAddr, err := parseIP(addr)
	if err != nil {
		return false, fmt.Errorf("unable to parse address: %s: %w", addr, err)
	}

	return checker.ContainsIP(ipAddr), nil
}

// ContainsIP checks if provided address is in the IPs.
func (checker *Checker) ContainsIP(addr net.IP) bool {
	for _, deniedIP := range checker.denyIPs {
		if deniedIP.Equal(addr) {
			return true
		}
	}

	for _, denyNet := range checker.denyIPsNet {
		if denyNet.Contains(addr) {
			return true
		}
	}

	return false
}

func parseIP(addr string) (net.IP, error) {
	userIP := net.ParseIP(addr)
	if userIP == nil {
		return nil, fmt.Errorf("unable parse IP from address %s", addr)
	}

	return userIP, nil
}
