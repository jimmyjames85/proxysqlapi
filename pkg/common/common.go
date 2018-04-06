package common

import "encoding/json"

type DBConfig struct {
	DBuser string `envconfig:"DB_USER" default:"root"`
	DBPswd string `envconfig:"DB_PASS" default:""`
	DBHost string `envconfig:"DB_HOST" default:"localhost"`
	DBPort int    `envconfig:"DB_PORT" default:"6032"`
}

func (c *DBConfig) ToJSON() string {
	copy := *c
	copy.DBPswd = "****"
	b, _ := json.Marshal(copy)
	return string(b)
}
