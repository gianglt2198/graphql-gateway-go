package config

type EtcdConfig struct {
	Endpoints []string `mapstructure:"endpoints" yaml:"endpoints" envDefault:"127.0.0.1:2379"`
	BasePath  string   `mapstructure:"base_path" yaml:"base_path" envDefault:"/services"`
}

type IntermediaryConfig struct {
	EtcdConfig `mapstructure:",squash"`
}
