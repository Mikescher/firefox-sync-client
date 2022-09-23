package models

import "time"

type Record struct {
	ID           string
	RawData      []byte
	Payload      string
	Modified     time.Time
	ModifiedUnix float64
	DecodedData  []byte
	SortIndex    int64
	TTL          *int64
}

type RecordUpdate struct {
	ID        string
	Payload   *string
	SortIndex *int64
	TTL       *int64
}
