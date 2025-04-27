package service

type GqlServiceConfig struct {
	Enabled     bool                     `yaml:"enabled,omitempty" envDefault:"false"`
	Debug       bool                     `yaml:"debug,omitempty" envDefault:"false"`
	GqlHttp     GqlHttpServiceConfig     `yaml:"http,omitempty"`
	GqlConsumer GqlConsumerServiceConfig `yaml:"consumer,omitempty"`
}

type GqlHttpServiceConfig struct {
	Enabled          bool                    `yaml:"enabled,omitempty" envDefault:"false"`
	Port             int                     `yaml:"port,omitempty" envDefault:"4000"`
	Path             string                  `yaml:"path,omitempty" envDefault:"/query"`
	SubscriptionPath string                  `yaml:"subscription_path,omitempty" envDefault:"/subscription"`
	Playground       GqlHttpPlayGroundConfig `yaml:"playground,omitempty"`
}

type GqlConsumerServiceConfig struct {
	Enabled bool `yaml:"enabled,omitempty" envDefault:"false"`
}
type GqlHttpPlayGroundConfig struct {
	Enabled bool   `yaml:"enabled,omitempty" envDefault:"false"`
	Path    string `yaml:"path,omitempty" envDefault:"/playground"`
}
