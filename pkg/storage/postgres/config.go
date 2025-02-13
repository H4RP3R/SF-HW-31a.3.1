package postgres

import "fmt"

type Config struct {
	User     string
	Password string
	Host     string
	Port     string
	DBName   string
}

func (c *Config) ConString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", c.User, c.Password, c.Host, c.Port, c.DBName)
}
