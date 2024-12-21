package configs

type Config struct{}

func MustConfig() (*Config, error) {
	var config Config

	return &config, nil
}
