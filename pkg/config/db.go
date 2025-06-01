package config

type DBConfig struct {
	Trace      bool   `mapstructures:"trace" default:"false"`
	Driver     string `mapstructures:"driver" default:"sqlite3"`
	Connection string `mapstructures:"connection" default:"file:ent?mode=memory&cache=shared&_fk=1"`
}
