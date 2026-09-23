package requestparamsmodel

import (
	"strings"
	"time"
)

type RequestParams struct {
	Origin        string  `form:"origin" json:"origin" binding:"required,len=3"`
	Destination   string  `form:"destination" json:"destination" binding:"required,len=3"`
	DepartureDate string  `form:"departureDate" json:"departureDate" binding:"required,datetime=2006-01-02"`
	ReturnDate    *string `form:"returnDate" json:"returnDate" binding:"omitempty,datetime=2006-01-02"`
	Passengers    int     `form:"passengers" json:"passengers" binding:"required,min=1"`
	CabinClass    string  `form:"cabinClass" json:"cabinClass" binding:"required"`
	Filters
}

type Filters struct {
	MinPrice           *int64   `form:"minPrice" json:"minPrice" binding:"omitempty,min=0"`
	MaxPrice           *int64   `form:"maxPrice" json:"maxPrice" binding:"omitempty,min=0"`
	MaxStops           *int     `form:"maxStops" json:"maxStops" binding:"omitempty,min=0"`
	MaxDurationMinutes *int     `form:"maxDuration" json:"maxDuration" binding:"omitempty,min=1"`
	Airlines           []string `form:"airlines" json:"airlines"` // IATA: QZ, ID, GA, JT
	DepartureFrom      string   `form:"departureFrom" json:"departureFrom" binding:"omitempty,datetime=15:04"`
	DepartureTo        string   `form:"departureTo" json:"departureTo" binding:"omitempty,datetime=15:04"`
	ArrivalFrom        string   `form:"arrivalFrom" json:"arrivalFrom" binding:"omitempty,datetime=15:04"`
	ArrivalTo          string   `form:"arrivalTo" json:"arrivalTo" binding:"omitempty,datetime=15:04"`
}

func (r *RequestParams) Normalize() {
	r.Origin = strings.ToUpper(strings.TrimSpace(r.Origin))
	r.Destination = strings.ToUpper(strings.TrimSpace(r.Destination))
	r.CabinClass = strings.ToLower(strings.TrimSpace(r.CabinClass))
	airlines := make([]string, 0, len(r.Airlines))
	for _, airline := range r.Airlines {
		if airline = strings.ToUpper(strings.TrimSpace(airline)); airline != "" {
			airlines = append(airlines, airline)
		}
	}
	r.DepartureFrom = normalizeClock(r.DepartureFrom)
	r.DepartureTo = normalizeClock(r.DepartureTo)
	r.ArrivalFrom = normalizeClock(r.ArrivalFrom)
	r.ArrivalTo = normalizeClock(r.ArrivalTo)
	r.Airlines = airlines
}

func normalizeClock(s string) string {
	t, err := time.Parse("15:04", strings.TrimSpace(s))
	if err != nil {
		return s
	}
	return t.Format("15:04")
}
