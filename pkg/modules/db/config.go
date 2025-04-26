package db

type Config struct {
	Enabled    bool   `yaml:"enabled" envDefault:"false"`
	Trace      bool   `yaml:"trace" envDefault:"false"`
	Driver     string `yaml:"driver" envDefault:"sqlite3"`
	Connection string `yaml:"connection" envDefault:"file:ent?mode=memory&cache=shared&_fk=1"`
}
