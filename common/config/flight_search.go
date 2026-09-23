package config

type FlightSearchConfig struct {
	LoadTimeout          Duration `yaml:"load_timeout"`
	CacheTTL             Duration `yaml:"cache_ttl"`
	CacheCleanupInterval Duration `yaml:"cache_cleanup_interval"`
}
