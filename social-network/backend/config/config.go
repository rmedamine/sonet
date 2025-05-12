package config

import (
	"database/sql"
)

const (
	DATABASE_NAME    = "social_network.db"
	DRIVER_NAME      = "sqlite3"
	ADDRS            = ":8000"
	TEMPLATE_DIR     = "./template"
	SESSION_EXP_TIME = 24 * 60 * 60 // in seconds
	LIMIT_PER_PAGE   = 5
)

var (
	DB      *sql.DB          = nil
	// TMPL    *TemplateManager = nil
	SESSION *SessionManager  = nil
)
