package config

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	HTTP       HTTPConfig       `mapstructure:"http"`
	GraphQL    GraphQLConfig    `mapstructure:"graphql"`
	Federation FederationConfig `mapstructure:"federation"`
}

type HTTPConfig struct {
	Enabled      bool `mapstructure:"enabled"`
	Port         int  `mapstructure:"port"`
	ReadTimeout  int  `mapstructure:"read_timeout"`
	WriteTimeout int  `mapstructure:"write_timeout"`
	IdleTimeout  int  `mapstructure:"idle_timeout"`
}

type GraphQLConfig struct {
	Enabled    bool             `mapstructure:"enabled"`
	Playground bool             `mapstructure:"playground"`
	Complexity ComplexityConfig `mapstructure:"complexity"`
}

type FederationConfig struct {
	Enabled    bool             `mapstructure:"enabled"`
	Playground bool             `mapstructure:"playground"`
	Subgraphs  []SubgraphConfig `mapstructure:"subgraphs" json:"subgraphs"`
	Complexity ComplexityConfig `mapstructure:"complexity"`
}

type SubgraphConfig struct {
	Name    string            `mapstructure:"name" json:"name"`
	URL     string            `mapstructure:"url" json:"url"`
	Headers map[string]string `mapstructure:"headers" json:"headers"`
	Timeout int               `mapstructure:"timeout" json:"timeout"`
	Retries int               `mapstructure:"retries" json:"retries"`
}

type ComplexityConfig struct {
	Enabled bool `yaml:"enabled,omitempty"`
	Limit   int  `yaml:"limit,omitempty"`
}
