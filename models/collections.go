package models

import "time"

type CollectionInfo struct {
	Name         string
	LastModified time.Time
}

type CollectionCount struct {
	Name  string
	Count int
}

type CollectionUsage struct {
	Name  string
	Usage int64 // bytes
}

type Collection struct {
	Name         string
	LastModified time.Time
	Count        int
	Usage        int64 // bytes
}
