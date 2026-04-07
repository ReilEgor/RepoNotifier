package config

type (
	RedisHostType     string
	RedisPortType     string
	RedisPasswordType string

	HTTPPortType string
	DSNType      string
)
type Config struct {
	DSN           DSNType           `env:"DB_SOURCE" envDefault:"postgres://user:password@localhost:5432/userservice?sslmode=disable"`
	HTTPPort      HTTPPortType      `env:"HTTP_PORT" envDefault:"8080"`
	RedisHost     RedisHostType     `env:"REDIS_HOST" envDefault:"redis"`
	RedisPort     RedisPortType     `env:"REDIS_PORT" envDefault:"6379"`
	RedisPassword RedisPasswordType `env:"REDIS_PASSWORD" envDefault:"redis_password"`
}
