package models

import (
	"encoding/xml"
	"ffsyncclient/cli"
	"ffsyncclient/langext"
	"time"
)

type PasswordPayloadSchema struct {
	ID                  string  `json:"id"`
	Hostname            string  `json:"hostname"`
	FormSubmitURL       string  `json:"formSubmitURL"`
	HTTPRealm           *string `json:"httpRealm"`
	Username            string  `json:"username"`
	Password            string  `json:"password"`
	UsernameField       string  `json:"usernameField"`
	PasswordField       string  `json:"passwordField"`
	TimeCreated         *int64  `json:"timeCreated,omitempty"`
	TimePasswordChanged *int64  `json:"timePasswordChanged,omitempty"`
	TimeLastUsed        *int64  `json:"timeLastUsed,omitempty"`
	TimesUsed           *int64  `json:"timesUsed,omitempty"`
}

func (j PasswordPayloadSchema) ToModel() PasswordRecord {
	return PasswordRecord{
		ID:              j.ID,
		Hostname:        j.Hostname,
		FormSubmitURL:   j.FormSubmitURL,
		HTTPRealm:       j.HTTPRealm,
		Username:        j.Username,
		Password:        j.Password,
		UsernameField:   j.UsernameField,
		PasswordField:   j.PasswordField,
		Created:         optMilliTime(j.TimeCreated),
		PasswordChanged: optMilliTime(j.TimePasswordChanged),
		LastUsed:        optMilliTime(j.TimeLastUsed),
		TimesUsed:       j.TimesUsed,
	}
}

func optMilliTime(ms *int64) *time.Time {
	if ms == nil || *ms == 0 {
		return nil
	} else {
		return langext.Ptr(time.UnixMilli(*ms))
	}
}

type PasswordRecord struct {
	ID              string
	Hostname        string
	FormSubmitURL   string
	HTTPRealm       *string
	Username        string
	Password        string
	UsernameField   string
	PasswordField   string
	Created         *time.Time
	PasswordChanged *time.Time
	LastUsed        *time.Time
	TimesUsed       *int64
}

func (pw PasswordRecord) ToJSON(ctx *cli.FFSContext, showPW bool) langext.H {
	return langext.H{
		"id":                   pw.ID,
		"hostname":             pw.Hostname,
		"formSubmitURL":        pw.FormSubmitURL,
		"httpRealm":            pw.HTTPRealm,
		"username":             pw.Username,
		"password":             fmtPass(pw.Password, showPW),
		"usernameField":        pw.UsernameField,
		"passwordField":        pw.PasswordField,
		"created":              fmOptDateToNullable(ctx, pw.Created),
		"created_unix":         fmOptDateToNullableUnix(pw.Created),
		"passwordChanged":      fmOptDateToNullable(ctx, pw.PasswordChanged),
		"passwordChanged_unix": fmOptDateToNullableUnix(pw.PasswordChanged),
		"lastUsed":             fmOptDateToNullable(ctx, pw.LastUsed),
		"lastUsed_unix":        fmOptDateToNullableUnix(pw.LastUsed),
		"timesUsed":            pw.TimesUsed,
	}
}

func (pw PasswordRecord) ToXML(ctx *cli.FFSContext, node string, showPW bool) any {
	type xmlentry struct {
		XMLName             xml.Name
		ID                  string  `xml:"ID,attr"`
		Hostname            string  `xml:"Hostname,attr"`
		FormSubmitURL       string  `xml:"FormSubmitURL,attr"`
		HTTPRealm           *string `xml:"HTTPRealm,omitempty,attr"`
		Username            string  `xml:"Username,attr"`
		Password            string  `xml:"Password,attr"`
		UsernameField       string  `xml:"UsernameField,attr"`
		PasswordField       string  `xml:"PasswordField,attr"`
		Created             *string `xml:"Created,omitempty,attr"`
		CreatedUnix         *int64  `xml:"CreatedUnix,omitempty,attr"`
		PasswordChanged     *string `xml:"PasswordChanged,omitempty,attr"`
		PasswordChangedUnix *int64  `xml:"PasswordChangedUnix,omitempty,attr"`
		LastUsed            *string `xml:"LastUsed,omitempty,attr"`
		LastUsedUnix        *int64  `xml:"LastUsedUnix,omitempty,attr"`
		TimesUsed           *int64  `xml:"TimesUsed,omitempty,attr"`
	}
	return xmlentry{
		XMLName:             xml.Name{Local: node},
		ID:                  pw.ID,
		Hostname:            pw.Hostname,
		FormSubmitURL:       pw.FormSubmitURL,
		HTTPRealm:           pw.HTTPRealm,
		Username:            pw.Username,
		Password:            fmtPass(pw.Password, showPW),
		UsernameField:       pw.UsernameField,
		PasswordField:       pw.PasswordField,
		Created:             fmOptDateToNullable(ctx, pw.Created),
		CreatedUnix:         fmOptDateToNullableUnix(pw.Created),
		PasswordChanged:     fmOptDateToNullable(ctx, pw.PasswordChanged),
		PasswordChangedUnix: fmOptDateToNullableUnix(pw.PasswordChanged),
		LastUsed:            fmOptDateToNullable(ctx, pw.LastUsed),
		LastUsedUnix:        fmOptDateToNullableUnix(pw.LastUsed),
		TimesUsed:           pw.TimesUsed,
	}
}

func (pw PasswordRecord) FormatPassword(showPW bool) string {
	return fmtPass(pw.Password, showPW)
}
