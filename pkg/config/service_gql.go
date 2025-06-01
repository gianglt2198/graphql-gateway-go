package config

type GqlServiceConfig struct {
	Enabled  bool                     `mapstructure:"enabled,omitempty" envDefault:"false"`
	Debug    bool                     `mapstructure:"debug,omitempty" envDefault:"false"`
	Http     GqlHttpServiceConfig     `mapstructure:"http,omitempty"`
	Consumer GqlConsumerServiceConfig `mapstructure:"consumer,omitempty"`
}

type GqlHttpServiceConfig struct {
	Enabled           bool   `mapstructure:"enabled,omitempty" envDefault:"false"`
	Port              int    `mapstructure:"port,omitempty" envDefault:"4000"`
	Path              string `mapstructure:"path,omitempty" envDefault:"/query"`
	SubscriptionPath  string `mapstructure:"subscription_path,omitempty" envDefault:"/subscription"`
	PlaygroundEnabled bool   `mapstructure:"playground_enabled,omitempty" envDefault:"false"`
}

type GqlConsumerServiceConfig struct {
	Enabled bool `mapstructure:"enabled,omitempty" envDefault:"false"`
}
