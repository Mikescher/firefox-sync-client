package cli

import (
	"ffsyncclient/consts"
)

type Options struct {
	Quiet          bool
	Verbose        bool
	Format         OutputFormat
	ConfigFilePath string
	ServerURL      string
}

func DefaultCLIOptions() Options {
	return Options{
		Quiet:          false,
		Verbose:        false,
		Format:         OutputFormatText,
		ConfigFilePath: "~/.config/firefox-sync-client.secret",
		ServerURL:      consts.ServerURLProduction,
	}
}
