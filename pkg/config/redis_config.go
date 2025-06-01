package config

type RedisConfig struct {
	URL        string `mapstructure:"url,omitempty" yaml:"url,omitempty" envDefault:"redis://127.0.0.1:6379"`
	Namespace  string `mapstructure:"namespace,omitempty" yaml:"namespace,omitempty"`
	TLSEnabled bool   `mapstructure:"tls_enabled" yaml:"tls_enabled" envDefault:"false"`
	ScanCount  int    `mapstructure:"scan_count" yaml:"scan_count" envDefault:"5000"`
}
