package common

import "encoding/json"

type DBConfig struct {
	DBuser string `envconfig:"ADMIN_USER" default:"root"`
	DBPswd string `envconfig:"ADMIN_PASS" default:""`
	DBHost string `envconfig:"ADMIN_HOST" default:"localhost"`
	DBPort int    `envconfig:"ADMIN_PORT" default:"6032"`
}

func (c *DBConfig) ToJSON() string {
	copy := *c
	copy.DBPswd = "****"
	b, _ := json.Marshal(copy)
	return string(b)
}
