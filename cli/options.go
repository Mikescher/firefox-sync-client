package cli

import (
	"ffsyncclient/consts"
	"gogs.mikescher.com/BlackForestBytes/goext/termext"
	"time"
)

type Options struct {
	Quiet                         bool
	Verbose                       bool
	Format                        *OutputFormat
	SessionFilePath               string
	AuthServerURL                 string
	TokenServerURL                string
	OutputColor                   bool
	OutputFile                    *string
	TimeZone                      *time.Location
	TimeFormat                    string
	SaveRefreshedSession          bool
	ForceRefreshSession           bool
	NoXMLDeclaration              bool
	LinearizeJson                 bool
	ManualAuthLoginEmail          *string
	ManualAuthLoginPassword       *string
	RequestX509RetryDelay         time.Duration
	RequestFloodControlRetryDelay time.Duration
	RequestServerErrRetryDelay    time.Duration
	MaxRequestRetries             int
	RequestTimeout                time.Duration
}

func DefaultCLIOptions() Options {
	return Options{
		Quiet:                         false,
		Verbose:                       false,
		Format:                        nil,
		SessionFilePath:               "~/.config/firefox-sync-client.secret",
		AuthServerURL:                 consts.ServerURLProduction,
		TokenServerURL:                consts.TokenServerURL,
		OutputColor:                   termext.SupportsColors(),
		TimeZone:                      time.Local,
		TimeFormat:                    "2006-01-02 15:04:05Z07:00",
		OutputFile:                    nil,
		SaveRefreshedSession:          true,
		ForceRefreshSession:           false,
		ManualAuthLoginEmail:          nil,
		ManualAuthLoginPassword:       nil,
		RequestX509RetryDelay:         5 * time.Second,
		RequestFloodControlRetryDelay: 15 * time.Second,
		RequestServerErrRetryDelay:    1 * time.Second,
		MaxRequestRetries:             5,
		RequestTimeout:                10 * time.Second,
	}
}
