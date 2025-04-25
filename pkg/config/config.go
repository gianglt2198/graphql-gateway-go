package config

type Config struct {
	Name string `yaml:"name,omitempty" envDefault:"Unknown"`
}
