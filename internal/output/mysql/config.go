package mysql

import cfg "github.com/sherifabdlnaby/prism/pkg/config"

//config struct
type config struct {
	Username string `mapstructure:"username" validate:"required"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"db_name" validate:"required"`
	Query    string `mapstructure:"query" validate:"required"`
	query    cfg.Selector
}
