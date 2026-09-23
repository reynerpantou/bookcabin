package aviation

// in production, use persistent database such as postgre / mysql / can simply store it under dynamic configuration
var airportCities = map[string]string{
	"CGK": "Jakarta",
	"DPS": "Denpasar",
	"SOC": "Solo",
	"SUB": "Surabaya",
	"UPG": "Makassar",
}

func GetCityFromAirportCode(code string) string {
	return airportCities[code]
}
