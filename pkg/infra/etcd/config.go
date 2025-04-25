package etcd

type Config struct {
	Enabled   bool     `yaml:"enabled" envDefault:"false"`
	Endpoints []string `yaml:"endpoints" envDefault:"127.0.0.1:2379"`
	BasePath  string   `yaml:"base_path" envDefault:"/services"`
}

type IntermediaryConfig struct {
	Config
}
