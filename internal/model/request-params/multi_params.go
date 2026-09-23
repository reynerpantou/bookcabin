package requestparamsmodel

import (
	"errors"
	"fmt"
)

type MultiRequestParams struct {
	Legs []RequestParams `json:"legs" binding:"required,min=2,max=6,dive"`
}

func (m *MultiRequestParams) Normalize() {
	for i := range m.Legs {
		m.Legs[i].Normalize()
	}
}

func (m *MultiRequestParams) Validate() error {
	for i := range m.Legs {
		leg := &m.Legs[i]
		if err := leg.Validate(); err != nil {
			return fmt.Errorf("leg %d: %w", i+1, err)
		}
		if leg.ReturnDate != nil && *leg.ReturnDate != "" {
			return fmt.Errorf("leg %d: returnDate is not allowed; add the return as its own leg", i+1)
		}
		if i == 0 {
			continue
		}
		prev := &m.Legs[i-1]
		if leg.DepartureDate < prev.DepartureDate {
			return fmt.Errorf("leg %d departs before leg %d", i+1, i)
		}
		if leg.Passengers != prev.Passengers {
			return errors.New("all legs must have the same number of passengers")
		}
	}
	return nil
}
