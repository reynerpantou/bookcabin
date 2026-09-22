package config

type AirlinesConfig struct {
	AirAsia         AirlineConfig `yaml:"airasia"`
	BatikAir        AirlineConfig `yaml:"batik_air"`
	GarudaIndonesia AirlineConfig `yaml:"garuda_indonesia"`
	LionAir         AirlineConfig `yaml:"lion_air"`
}

type AirlineConfig struct {
	Enabled bool       `yaml:"enabled"`
	Timeout string     `yaml:"timeout"`
	Mock    MockConfig `yaml:"mock"`
}

type MockConfig struct {
	FilePath    string   `yaml:"file_path"`
	MinDelay    string   `yaml:"min_delay"`
	MaxDelay    string   `yaml:"max_delay"`
	SuccessRate *float64 `yaml:"success_rate"`
}
