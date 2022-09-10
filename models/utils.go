package models

import (
	"encoding/json"
	"ffsyncclient/cli"
	"ffsyncclient/langext"
	"fmt"
	"github.com/joomcode/errorx"
	"time"
)

func UnmarshalPasswords(ctx *cli.FFSContext, records []Record, ignoreSchemaErrors bool) ([]PasswordRecord, error) {
	result := make([]PasswordRecord, 0, len(records))

	for _, v := range records {
		var jsonschema PasswordPayloadSchema
		err := json.Unmarshal(v.DecodedData, &jsonschema)
		if err != nil {
			if ignoreSchemaErrors {
				ctx.PrintVerbose(fmt.Sprintf("Failed to decode record %s to password-schema -- skipping", v.ID))
				continue
			}

			return nil, errorx.Decorate(err, fmt.Sprintf("Failed to decode record %s to password-schema\n%s", v.ID, string(v.DecodedData)))
		}

		result = append(result, jsonschema.ToModel())

		ctx.PrintVerbose(fmt.Sprintf("Decoded record %s (%s)", v.ID, jsonschema.Hostname))
	}

	return result, nil
}

func UnmarshalPassword(ctx *cli.FFSContext, record Record) (PasswordRecord, error) {
	var jsonschema PasswordPayloadSchema
	err := json.Unmarshal(record.DecodedData, &jsonschema)
	if err != nil {
		return PasswordRecord{}, errorx.Decorate(err, fmt.Sprintf("Failed to decode record %s to password-schema\n%s", record.ID, string(record.DecodedData)))
	}

	model := jsonschema.ToModel()

	ctx.PrintVerbose(fmt.Sprintf("Decoded record %s (%s)", record.ID, jsonschema.Hostname))

	return model, nil
}

func UnmarshalBookmarks(ctx *cli.FFSContext, records []Record, ignoreSchemaErrors bool) ([]BookmarkRecord, error) {
	result := make([]BookmarkRecord, 0, len(records))

	for _, v := range records {
		var jsonschema BookmarkPayloadSchema
		err := json.Unmarshal(v.DecodedData, &jsonschema)
		if err != nil {
			if ignoreSchemaErrors {
				ctx.PrintVerbose(fmt.Sprintf("Failed to decode record %s to bookmark-schema -- skipping", v.ID))
				continue
			}

			return nil, errorx.Decorate(err, fmt.Sprintf("Failed to decode record %s to bookmark-schema\n%s", v.ID, string(v.DecodedData)))
		}

		result = append(result, jsonschema.ToModel())

		ctx.PrintVerbose(fmt.Sprintf("Decoded record %s (%s)", v.ID, jsonschema.Title))
	}

	return result, nil
}

func UnmarshalBookmark(ctx *cli.FFSContext, record Record) (BookmarkRecord, error) {
	var jsonschema BookmarkPayloadSchema
	err := json.Unmarshal(record.DecodedData, &jsonschema)
	if err != nil {
		return BookmarkRecord{}, errorx.Decorate(err, fmt.Sprintf("Failed to decode record %s to bookmark-schema\n%s", record.ID, string(record.DecodedData)))
	}

	model := jsonschema.ToModel()

	ctx.PrintVerbose(fmt.Sprintf("Decoded record %s (%s)", record.ID, jsonschema.Title))

	return model, nil
}

func fmtPass(pw string, showPW bool) string {
	if showPW {
		return pw
	} else {
		return "***"
	}
}

func fmOptDate(ctx *cli.FFSContext, d *time.Time) string {
	if d == nil {
		return ""
	}
	return d.In(ctx.Opt.TimeZone).Format(ctx.Opt.TimeFormat)
}

func fmOptDateToNullable(ctx *cli.FFSContext, d *time.Time) *string {
	if d == nil {
		return nil
	}
	return langext.Ptr(d.In(ctx.Opt.TimeZone).Format(ctx.Opt.TimeFormat))
}

func fmOptDateToNullableUnix(d *time.Time) *int64 {
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
