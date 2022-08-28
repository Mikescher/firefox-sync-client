package models

import "time"

type DecodedRecord struct {
	ID          string
	Payload     string
	Modified    time.Time
	DecodedData []byte
}
