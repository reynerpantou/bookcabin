package requestparamsmodel

type RequestParams struct {
	Origin        string  `form:"origin" json:"origin" binding:"required"`
	Destination   string  `form:"destination" json:"destination" binding:"required"`
	DepartureDate string  `form:"departureDate" json:"departureDate" binding:"required"`
	ReturnDate    *string `form:"returnDate" json:"returnDate"`
	Passengers    int     `form:"passengers" json:"passengers" binding:"required,min=1"`
	CabinClass    string  `form:"cabinClass" json:"cabinClass" binding:"required"`
}
