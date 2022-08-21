package cli

import (
	"ffsyncclient/consts"
)

type Options struct {
	Quiet          bool
	Verbose        bool
	Format         OutputFormat
	ConfigFilePath string
	AuthServerURL  string
	TokenServerURL string
	OutputColor    *bool
}

func DefaultCLIOptions() Options {
	return Options{
		Quiet:          false,
		Verbose:        false,
		Format:         OutputFormatText,
		ConfigFilePath: "~/.config/firefox-sync-client.secret",
		AuthServerURL:  consts.ServerURLProduction,
		TokenServerURL: consts.TokenServerURL,
		OutputColor:    nil,
	}
}
