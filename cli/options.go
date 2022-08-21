package cli

import (
	"ffsyncclient/consts"
	"time"
)

type Options struct {
	Quiet          bool
	Verbose        bool
	Format         *OutputFormat
	ConfigFilePath string
	AuthServerURL  string
	TokenServerURL string
	OutputColor    *bool
	TimeZone       *time.Location
	TimeFormat     string
}

func DefaultCLIOptions() Options {
	return Options{
		Quiet:          false,
		Verbose:        false,
		Format:         nil,
		ConfigFilePath: "~/.config/firefox-sync-client.secret",
		AuthServerURL:  consts.ServerURLProduction,
		TokenServerURL: consts.TokenServerURL,
		OutputColor:    nil,
		TimeZone:       time.Local,
		TimeFormat:     "2006-01-02 15:04:05Z07:00",
	}
}
