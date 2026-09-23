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

// Power Outlet -> power_outlet
// Meal -> meal
// wifi -> wifi
func NormalizeAmenities(amenities []string) []string {
	out := make([]string, 0, len(amenities))
	for _, a := range amenities {
		words := strings.Fields(strings.ToLower(a))
		if len(words) == 0 {
			continue
		}
		out = append(out, strings.Join(words, "_"))
	}
	return out
}
