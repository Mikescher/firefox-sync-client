package models

import (
	"ffsyncclient/langext"
	"time"
)

type passwordRecordJson struct {
	ID                  string  `json:"id"`
	Hostname            string  `json:"hostname"`
	FormSubmitURL       string  `json:"formSubmitURL"`
	HTTPRealm           *string `json:"httpRealm"`
	Username            string  `json:"username"`
	Password            string  `json:"password"`
	UsernameField       string  `json:"usernameField"`
	PasswordField       string  `json:"passwordField"`
	TimeCreated         *int64  `json:"timeCreated"`
	TimePasswordChanged *int64  `json:"timePasswordChanged"`
	TimeLastUsed        *int64  `json:"timeLastUsed"`
	TimesUsed           *int64  `json:"timesUsed"`
}

func (j passwordRecordJson) ToModel() PasswordRecord {
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
