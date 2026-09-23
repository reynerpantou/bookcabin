package response

type FlightMultiSearchResponse struct {
	Legs []FlightSearchResponse `json:"legs"`
}
