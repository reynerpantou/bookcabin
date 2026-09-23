package aviation

type Provider string

const (
	AirAsiaProvider         Provider = "AirAsia"
	BatikAirProvider        Provider = "Batik Air"
	GarudaIndonesiaProvider Provider = "Garuda Indonesia"
	LionAirProvider         Provider = "Lion Air"
)

func (p Provider) String() string {
	return string(p)
}
