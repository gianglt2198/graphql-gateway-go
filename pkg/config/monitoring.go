package config

type MonitoringConfig struct {
	Enabled  bool   `yaml:"enabled,omitempty" envDefault:"false"`
	Endpoint string `yaml:"endpoint,omitempty" envDefault:"localhost:4317"`
}
