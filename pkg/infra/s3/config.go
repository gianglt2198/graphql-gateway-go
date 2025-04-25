package s3

type S3Config struct {
	Enabled bool   `yaml:"enabled" envDefault:"false"`
	Region  string `yaml:"region,omitempty"`
	Bucket  string `yaml:"bucket,omitempty"`
}
