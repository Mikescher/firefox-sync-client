package models

import (
	"encoding/json"
	"encoding/xml"
	"ffsyncclient/cli"
	"github.com/joomcode/errorx"
	"gogs.mikescher.com/BlackForestBytes/goext/langext"
	"time"
)

type PasswordPayloadSchema struct {
	ID                  string  `json:"id"`
	Deleted             bool    `json:"deleted,omitempty"`
	Hostname            string  `json:"hostname"`
	FormSubmitURL       string  `json:"formSubmitURL"`
	HTTPRealm           *string `json:"httpRealm,omitempty"`
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
		Deleted:         j.Deleted,
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

type PasswordRecord struct {
	ID              string
	Deleted         bool
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
		"deleted":              pw.Deleted,
		"password":             fmtPass(pw.Password, showPW),
		"usernameField":        pw.UsernameField,
		"passwordField":        pw.PasswordField,
		"created":              fmtOptDateToNullable(ctx, pw.Created),
		"created_unix":         fmtOptDateToNullableUnix(pw.Created),
		"passwordChanged":      fmtOptDateToNullable(ctx, pw.PasswordChanged),
		"passwordChanged_unix": fmtOptDateToNullableUnix(pw.PasswordChanged),
		"lastUsed":             fmtOptDateToNullable(ctx, pw.LastUsed),
		"lastUsed_unix":        fmtOptDateToNullableUnix(pw.LastUsed),
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
		Created:             fmtOptDateToNullable(ctx, pw.Created),
		CreatedUnix:         fmtOptDateToNullableUnix(pw.Created),
		PasswordChanged:     fmtOptDateToNullable(ctx, pw.PasswordChanged),
		PasswordChangedUnix: fmtOptDateToNullableUnix(pw.PasswordChanged),
		LastUsed:            fmtOptDateToNullable(ctx, pw.LastUsed),
		LastUsedUnix:        fmtOptDateToNullableUnix(pw.LastUsed),
		TimesUsed:           pw.TimesUsed,
	}
}

func (pw PasswordRecord) ToPlaintextPayload() (string, error) {

	var created *int64 = nil
	if pw.Created != nil {
		created = langext.Ptr(pw.Created.UnixMilli())
	}

	var passwordChanged *int64 = nil
	if pw.PasswordChanged != nil {
		passwordChanged = langext.Ptr(pw.PasswordChanged.UnixMilli())
	}

	var lastUsed *int64 = nil
	if pw.LastUsed != nil {
		lastUsed = langext.Ptr(pw.LastUsed.UnixMilli())
	}

	obj := PasswordPayloadSchema{
		ID:                  pw.ID,
		Hostname:            pw.Hostname,
		FormSubmitURL:       pw.FormSubmitURL,
		HTTPRealm:           pw.HTTPRealm,
		Username:            pw.Username,
		Password:            pw.Password,
		UsernameField:       pw.UsernameField,
		PasswordField:       pw.PasswordField,
		TimeCreated:         created,
		TimePasswordChanged: passwordChanged,
		TimeLastUsed:        lastUsed,
		TimesUsed:           pw.TimesUsed,
	}

	pp, err := json.Marshal(obj)
	if err != nil {
		return "", errorx.Decorate(err, "failed to marshal password payload")
	}
	return string(pp), nil
}

func (pw PasswordRecord) FormatPassword(showPW bool) string {
	return fmtPass(pw.Password, showPW)
}
