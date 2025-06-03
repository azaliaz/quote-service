package rest

type Config struct {
	Port uint64 `env:"PORT" yaml:"port"`
}
