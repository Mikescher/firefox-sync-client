package models

import (
	"encoding/json"
	"ffsyncclient/cli"
	"fmt"
	"github.com/joomcode/errorx"
)

func ParsePasswords(ctx *cli.FFSContext, records []Record, ignoreSchemaErrors bool) ([]PasswordRecord, error) {
	result := make([]PasswordRecord, 0, len(records))

	for _, v := range records {
		var jsonschema passwordRecordJson
		err := json.Unmarshal(v.DecodedData, &jsonschema)
		if err != nil {
			if ignoreSchemaErrors {
				ctx.PrintVerbose(fmt.Sprintf("Failed to decode record %s to password-schema -- skipping", v.ID))
				continue
			}

			return nil, errorx.Decorate(err, fmt.Sprintf("Failed to decode record %s to password-schema\n%s", v.ID, string(v.DecodedData)))
		}

		result = append(result, jsonschema.ToModel())
	}

	return result, nil
}
