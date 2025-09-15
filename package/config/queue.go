package config

type QueueConfig struct {
	Enabled   bool             `mapstructure:"enabled" yaml:"enabled"`
	Providers []ProviderConfig `mapstructure:"providers" yaml:"providers"`
}

type ProviderConfig struct {
	Priorities  []PriorityConfig `mapstructure:"priorities" yaml:"priorities"`
	Concurrency int              `mapstructure:"concurrency" yaml:"concurrency"`
}

type PriorityConfig struct {
	Name     string `mapstructure:"name" yaml:"name"`
	Priority int    `mapstructure:"priority" yaml:"priority"`
}
