package models

import "time"

type Record struct {
	ID          string
	RawData     []byte
	Payload     string
	Modified    time.Time
	DecodedData []byte
	SortIndex   int64
	Deleted     bool
	TTL         *int64
}

type RecordUpdate struct {
	ID        string
	Payload   *string
	SortIndex *int64
	Deleted   *bool
	TTL       *int64
}
