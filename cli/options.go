package cli

type Options struct {
	Quiet          bool
	Verbose        bool
	Format         OutputFormat
	ConfigFilePath string
}

func DefaultCLIOptions() Options {
	return Options{
		Quiet:          false,
		Verbose:        false,
		Format:         OutputFormatText,
		ConfigFilePath: "~/.config/firefox-sync-client.secret",
	}
}
