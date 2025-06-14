package models

import (
	"encoding/json"
	"ffsyncclient/cli"
	"ffsyncclient/fferr"
	"fmt"
	"github.com/joomcode/errorx"
	"git.blackforestbytes.com/BlackForestBytes/goext/langext"
	"time"
)

func UnmarshalPasswords(ctx *cli.FFSContext, records []Record, ignoreSchemaErrors bool) ([]PasswordRecord, error) {
	result := make([]PasswordRecord, 0, len(records))

	for _, v := range records {
		if isAnonDeleted(v) {
			ctx.PrintVerbose(fmt.Sprintf("Record %s is deleted and no longer has any real payload - must skip", v.ID))
			continue
		}

		var jsonschema PasswordPayloadSchema
		err := json.Unmarshal(v.DecodedData, &jsonschema)
		err = checkIdFallthrough(err, v.ID, jsonschema.ID)
		if err != nil {
			if ignoreSchemaErrors {
				ctx.PrintVerbose(fmt.Sprintf("Failed to decode record %s to password-schema -- skipping", v.ID))
				continue
			}

			return nil, errorx.
				Decorate(err, fmt.Sprintf("Failed to decode record %s to password-schema", v.ID)).
				WithProperty(fferr.ExtraData, string(v.DecodedData))
		}

		result = append(result, jsonschema.ToModel())

		ctx.PrintVerbose(fmt.Sprintf("Decoded record %s (%s)", v.ID, jsonschema.Hostname))
	}

	return result, nil
}

func UnmarshalPassword(ctx *cli.FFSContext, record Record) (PasswordRecord, error) {
	var jsonschema PasswordPayloadSchema
	err := json.Unmarshal(record.DecodedData, &jsonschema)
	err = checkIdFallthrough(err, record.ID, jsonschema.ID)
	if err != nil {
		return PasswordRecord{}, errorx.
			Decorate(err, fmt.Sprintf("Failed to decode record %s to password-schema", record.ID)).
			WithProperty(fferr.ExtraData, string(record.DecodedData))
	}

	model := jsonschema.ToModel()

	ctx.PrintVerbose(fmt.Sprintf("Decoded record %s (%s)", record.ID, jsonschema.Hostname))

	return model, nil
}

func UnmarshalBookmarks(ctx *cli.FFSContext, records []Record, ignoreSchemaErrors bool) ([]BookmarkRecord, error) {
	result := make([]BookmarkRecord, 0, len(records))

	for _, v := range records {
		if isAnonDeleted(v) {
			ctx.PrintVerbose(fmt.Sprintf("Record %s is deleted and no longer has any real payload - must skip", v.ID))
			continue
		}

		var jsonschema BookmarkPayloadSchema
		err := json.Unmarshal(v.DecodedData, &jsonschema)
		err = checkIdFallthrough(err, v.ID, jsonschema.ID)
		if err != nil {
			if ignoreSchemaErrors {
				ctx.PrintVerbose(fmt.Sprintf("Failed to decode record %s to bookmark-schema -- skipping", v.ID))
				continue
			}

			return nil, errorx.
				Decorate(err, fmt.Sprintf("Failed to decode record %s to bookmark-schema", v.ID)).
				WithProperty(fferr.ExtraData, string(v.DecodedData))
		}

		result = append(result, jsonschema.ToModel())

		ctx.PrintVerbose(fmt.Sprintf("Decoded record %s (%s)", v.ID, jsonschema.Title))
	}

	return result, nil
}

func UnmarshalBookmark(ctx *cli.FFSContext, record Record) (BookmarkRecord, error) {
	var jsonschema BookmarkPayloadSchema
	err := json.Unmarshal(record.DecodedData, &jsonschema)
	err = checkIdFallthrough(err, record.ID, jsonschema.ID)
	if err != nil {
		return BookmarkRecord{}, errorx.
			Decorate(err, fmt.Sprintf("Failed to decode record %s to bookmark-schema", record.ID)).
			WithProperty(fferr.ExtraData, string(record.DecodedData))
	}

	model := jsonschema.ToModel()

	ctx.PrintVerbose(fmt.Sprintf("Decoded record %s (%s)", record.ID, jsonschema.Title))

	return model, nil
}

func UnmarshalForms(ctx *cli.FFSContext, records []Record, ignoreSchemaErrors bool) ([]FormRecord, error) {
	result := make([]FormRecord, 0, len(records))

	for _, v := range records {
		if isAnonDeleted(v) {
			ctx.PrintVerbose(fmt.Sprintf("Record %s is deleted and no longer has any real payload - must skip", v.ID))
			continue
		}

		var jsonschema FormPayloadSchema
		err := json.Unmarshal(v.DecodedData, &jsonschema)
		err = checkIdFallthrough(err, v.ID, jsonschema.ID)
		if err != nil {
			if ignoreSchemaErrors {
				ctx.PrintVerbose(fmt.Sprintf("Failed to decode record %s to form-schema -- skipping", v.ID))
				continue
			}

			return nil, errorx.
				Decorate(err, fmt.Sprintf("Failed to decode record %s to form-schema", v.ID)).
				WithProperty(fferr.ExtraData, string(v.DecodedData))
		}

		result = append(result, jsonschema.ToModel(v))

		ctx.PrintVerbose(fmt.Sprintf("Decoded record %s (%s)", v.ID, jsonschema.Name))
	}

	return result, nil
}

func UnmarshalForm(ctx *cli.FFSContext, record Record) (FormRecord, error) {
	var jsonschema FormPayloadSchema
	err := json.Unmarshal(record.DecodedData, &jsonschema)
	err = checkIdFallthrough(err, record.ID, jsonschema.ID)
	if err != nil {
		return FormRecord{}, errorx.
			Decorate(err, fmt.Sprintf("Failed to decode record %s to form-schema", record.ID)).
			WithProperty(fferr.ExtraData, string(record.DecodedData))
	}

	model := jsonschema.ToModel(record)

	ctx.PrintVerbose(fmt.Sprintf("Decoded record %s (%s)", record.ID, jsonschema.Name))

	return model, nil
}

func UnmarshalHistories(ctx *cli.FFSContext, records []Record, ignoreSchemaErrors bool) ([]HistoryRecord, error) {
	result := make([]HistoryRecord, 0, len(records))

	for _, v := range records {
		if isAnonDeleted(v) {
			ctx.PrintVerbose(fmt.Sprintf("Record %s is deleted and no longer has any real payload - must skip", v.ID))
			continue
		}

		var jsonschema HistoryPayloadSchema
		err := json.Unmarshal(v.DecodedData, &jsonschema)
		err = checkIdFallthrough(err, v.ID, jsonschema.ID)
		if err != nil {
			if ignoreSchemaErrors {
				ctx.PrintVerbose(fmt.Sprintf("Failed to decode record %s to history-schema -- skipping", v.ID))
				continue
			}

			return nil, errorx.
				Decorate(err, fmt.Sprintf("Failed to decode record %s to history-schema", v.ID)).
				WithProperty(fferr.ExtraData, string(v.DecodedData))
		}

		result = append(result, jsonschema.ToModel())

		ctx.PrintVerbose(fmt.Sprintf("Decoded record %s (%s)", v.ID, jsonschema.URI))
	}

	return result, nil
}

func UnmarshalHistory(ctx *cli.FFSContext, record Record) (HistoryRecord, error) {
	var jsonschema HistoryPayloadSchema
	err := json.Unmarshal(record.DecodedData, &jsonschema)
	err = checkIdFallthrough(err, record.ID, jsonschema.ID)
	if err != nil {
		return HistoryRecord{}, errorx.
			Decorate(err, fmt.Sprintf("Failed to decode record %s to history-schema", record.ID)).
			WithProperty(fferr.ExtraData, string(record.DecodedData))
	}

	model := jsonschema.ToModel()

	ctx.PrintVerbose(fmt.Sprintf("Decoded record %s (%s)", record.ID, jsonschema.URI))

	return model, nil
}

func UnmarshalTabs(ctx *cli.FFSContext, records []Record, ignoreSchemaErrors bool) ([]TabClientRecord, error) {
	result := make([]TabClientRecord, 0, len(records))

	for _, v := range records {
		if isAnonDeleted(v) {
			ctx.PrintVerbose(fmt.Sprintf("Record %s is deleted and no longer has any real payload - must skip", v.ID))
			continue
		}

		var jsonschema TabPayloadSchema
		err := json.Unmarshal(v.DecodedData, &jsonschema)
		err = checkIdFallthrough(err, v.ID, jsonschema.ID)
		if err != nil {
			if ignoreSchemaErrors {
				ctx.PrintVerbose(fmt.Sprintf("Failed to decode record %s to tab-schema -- skipping", v.ID))
				continue
			}

			return nil, errorx.
				Decorate(err, fmt.Sprintf("Failed to decode record %s to tab-schema", v.ID)).
				WithProperty(fferr.ExtraData, string(v.DecodedData))
		}

		result = append(result, jsonschema.ToClientModel())

		ctx.PrintVerbose(fmt.Sprintf("Decoded record %s (%s)", v.ID, jsonschema.ClientName))
	}

	return result, nil
}

func UnmarshalTab(ctx *cli.FFSContext, record Record) (TabClientRecord, error) {
	var jsonschema TabPayloadSchema
	err := json.Unmarshal(record.DecodedData, &jsonschema)
	err = checkIdFallthrough(err, record.ID, jsonschema.ID)
	if err != nil {
		return TabClientRecord{}, errorx.
			Decorate(err, fmt.Sprintf("Failed to decode record %s to tab-schema", record.ID)).
			WithProperty(fferr.ExtraData, string(record.DecodedData))
	}

	model := jsonschema.ToClientModel()

	ctx.PrintVerbose(fmt.Sprintf("Decoded record %s (%s)", record.ID, jsonschema.ClientName))

	return model, nil
}

func checkIdFallthrough(err error, id1 string, id2 string) error {
	if err != nil {
		return err
	}
	if id1 != id2 {
		return fferr.UnmarshalConsistency.New("cannot unmarshal, inner ID <> outer ID")
	}
	return nil
}

// isAnonDeleted returns true if (and-only-if) r contains a single json field deleted:true
// aka `{"deleted":true}`
// these are fully deleted records, which no longer contain any data, and are only here for deletion-syncing
func isAnonDeleted(r Record) bool {
	var jsonschema map[string]any
	err := json.Unmarshal(r.DecodedData, &jsonschema)
	if err != nil {
		return false
	}

	if len(jsonschema) != 1 {
		return false
	}

	delAny, ok := jsonschema["deleted"]
	if !ok {
		return false
	}

	del, ok := delAny.(bool)
	if !ok {
		return false
	}

	if !del {
		return false
	}

	return true
}

func fmtPass(pw string, showPW bool) string {
	if showPW {
		return pw
	} else {
		return "***"
	}
}

func fmtOptDate(ctx *cli.FFSContext, d *time.Time) string {
	if d == nil {
		return ""
	}
	return d.In(ctx.Opt.TimeZone).Format(ctx.Opt.TimeFormat)
}

func fmtOptDateToNullable(ctx *cli.FFSContext, d *time.Time) *string {
	if d == nil {
		return nil
	}
	return langext.Ptr(d.In(ctx.Opt.TimeZone).Format(ctx.Opt.TimeFormat))
}

func fmtOptDateToNullableUnix(d *time.Time) *int64 {
	if d == nil {
		return nil
	}
	return langext.Ptr(d.Unix())
}

func optMilliTime(ms *int64) *time.Time {
	if ms == nil || *ms == 0 {
		return nil
	} else {
		return langext.Ptr(time.UnixMilli(*ms))
	}
}
