package models

import (
	"ffsyncclient/cli"
	"ffsyncclient/langext"
	"time"
)

type BookmarkPayloadSchema struct {
	ID         string `json:"id"`                // [common]
	Deleted    bool   `json:"deleted,omitempty"` // [common]
	Type       string `json:"type"`              // [common]
	DateAdded  *int64 `json:"dateAdded"`         // [common]
	ParentID   string `json:"parentid"`          // [common]
	ParentName string `json:"parentName"`        // [common]

	Title             string   `json:"title"`         // [bookmark, microsummary, query, livemark, folder]
	URI               string   `json:"bmkUri"`        // [bookmark, microsummary, query]
	Description       string   `json:"description"`   // [bookmark, microsummary, query]
	LoadInSidebar     bool     `json:"loadInSidebar"` // [bookmark, microsummary, query]
	Tags              []string `json:"tags"`          // [bookmark, microsummary, query]
	Keyword           string   `json:"keyword"`       // [bookmark, microsummary, query]
	Children          []string `json:"children"`      // [folder, livemark]
	GeneratorUri      string   `json:"generatorUri"`  // [microsummary]
	StaticTitle       string   `json:"staticTitle"`   // [microsummary]
	FolderName        string   `json:"folderName"`    // [query]
	QueryID           string   `json:"queryId"`       // [query]
	SeparatorPosition int      `json:"pos"`           // [separator]

	HasDupe bool `json:"hasDupe"` // ??
}

func (j BookmarkPayloadSchema) ToModel() BookmarkRecord {
	return BookmarkRecord{
		ID:                j.ID,
		Deleted:           j.Deleted,
		Title:             j.Title,
		URI:               j.URI,
		Description:       j.Description,
		LoadInSidebar:     j.LoadInSidebar,
		Tags:              j.Tags,
		Keyword:           j.Keyword,
		ParentID:          j.ParentID,
		ParentName:        j.ParentName,
		Children:          j.Children,
		Type:              BookmarkType(j.Type),
		GeneratorUri:      j.GeneratorUri,
		StaticTitle:       j.StaticTitle,
		FolderName:        j.FolderName,
		QueryID:           j.QueryID,
		SeparatorPosition: j.SeparatorPosition,
		DateAdded:         optMilliTime(j.DateAdded),
	}
}

type BookmarkType string

const (
	BookmarkTypeBookmark     BookmarkType = "bookmark"
	BookmarkTypeMicroSummary BookmarkType = "microsummary"
	BookmarkTypeQuery        BookmarkType = "query"
	BookmarkTypeFolder       BookmarkType = "folder"
	BookmarkTypeLivemark     BookmarkType = "livemark"
	BookmarkTypeSeparator    BookmarkType = "separator"
)

type BookmarkRecord struct {
	ID                string
	Deleted           bool
	Title             string
	URI               string
	Description       string
	LoadInSidebar     bool
	Tags              []string
	Keyword           string
	ParentID          string
	ParentName        string
	Children          []string
	Type              BookmarkType
	GeneratorUri      string
	StaticTitle       string
	FolderName        string
	QueryID           string
	SeparatorPosition int
	DateAdded         *time.Time
}

func (bm BookmarkRecord) ToJSON(ctx *cli.FFSContext) langext.H {
	r := langext.H{
		"id":         bm.ID,
		"type":       bm.Type,
		"deleted":    bm.Deleted,
		"added":      fmOptDateToNullable(ctx, bm.DateAdded),
		"added_unix": fmOptDateToNullableUnix(bm.DateAdded),
	}

	switch bm.Type {
	case BookmarkTypeBookmark:
		r["title"] = bm.Title
		r["uri"] = bm.URI
		r["description"] = bm.Description
		r["loadInSidebar"] = bm.LoadInSidebar
		r["tags"] = langext.ForceArray(bm.Tags)
		r["keyword"] = bm.Keyword
		return r
	case BookmarkTypeMicroSummary:
		r["title"] = bm.Title
		r["uri"] = bm.URI
		r["description"] = bm.Description
		r["loadInSidebar"] = bm.LoadInSidebar
		r["tags"] = langext.ForceArray(bm.Tags)
		r["keyword"] = bm.Keyword
		r["generatorUri"] = bm.GeneratorUri
		r["staticTitle"] = bm.StaticTitle
		return r
	case BookmarkTypeQuery:
		r["title"] = bm.Title
		r["uri"] = bm.URI
		r["description"] = bm.Description
		r["loadInSidebar"] = bm.LoadInSidebar
		r["tags"] = langext.ForceArray(bm.Tags)
		r["keyword"] = bm.Keyword
		r["folderName"] = bm.FolderName
		r["queryId"] = bm.QueryID
		return r
	case BookmarkTypeFolder:
		r["title"] = bm.Title
		r["children"] = langext.ForceArray(bm.Children)
		return r
	case BookmarkTypeLivemark:
		r["title"] = bm.Title
		r["children"] = langext.ForceArray(bm.Children)
		return r
	case BookmarkTypeSeparator:
		r["pos"] = bm.SeparatorPosition
		return r
	default:
		return r
	}
}

func (bm BookmarkRecord) ToXML(ctx *cli.FFSContext, node string) any {
	return nil //TODO
}

func (bm BookmarkRecord) ToPlaintextPayload() (string, error) {
	return "", nil //TODO
}

type BookmarkTreeRecord struct {
	BookmarkRecord
	ResolvedChildren []*BookmarkTreeRecord
}

func (bmt BookmarkTreeRecord) ToTreeJSON(ctx *cli.FFSContext) langext.H {
	base := bmt.ToJSON(ctx)
	if bmt.Type == BookmarkTypeFolder || bmt.Type == BookmarkTypeLivemark {
		arr := make([]langext.H, 0, len(bmt.ResolvedChildren))
		for _, child := range bmt.ResolvedChildren {
			arr = append(arr, child.ToTreeJSON(ctx))
		}
		base["children"] = arr
	}
	return base
}
