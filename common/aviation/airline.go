package aviation

import (
	"fmt"
	"strings"
)

func GetAirlineIATACodeFromFlightCode(flightCode string) string {
	flightCode = strings.TrimSpace(flightCode)
	if len(flightCode) < 2 {
		return ""
	}
	return flightCode[0:2]
}

func NormalizeCabinClass(s string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "economy", "y":
		return "economy", nil
	default:
		return "", fmt.Errorf("unknown cabin class %q", s)
	}
}
