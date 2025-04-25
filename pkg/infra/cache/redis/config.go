package credis

type Config struct {
	Enabled    bool   `yaml:"enabled" envDefault:"false"`
	URL        string `yaml:"url,omitempty" envDefault:"redis://127.0.0.1:6379"`
	Namespace  string `yaml:"namespace,omitempty"`
	TLSEnabled bool   `yaml:"tls_enabled" envDefault:"false"`
	ScanCount  int    `yaml:"scan_count" envDefault:"5000"`
}
