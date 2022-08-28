package models

import "time"

type Record struct {
	ID          string
	Payload     string
	Modified    time.Time
	DecodedData []byte
}
