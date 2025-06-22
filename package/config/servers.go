package config

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	HTTP    HTTPConfig    `mapstructure:"http"`
	GraphQL GraphQLConfig `mapstructure:"graphql"`
}

type HTTPConfig struct {
	Enabled      bool `mapstructure:"enabled"`
	Port         int  `mapstructure:"port"`
	ReadTimeout  int  `mapstructure:"read_timeout"`
	WriteTimeout int  `mapstructure:"write_timeout"`
	IdleTimeout  int  `mapstructure:"idle_timeout"`
}

type GraphQLConfig struct {
	Enabled    bool `mapstructure:"enabled"`
	Playground bool `mapstructure:"playground"`
}
