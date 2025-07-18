package models

import (
	"encoding/xml"
	"ffsyncclient/cli"
	"git.blackforestbytes.com/BlackForestBytes/goext/langext"
	"strings"
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
	FeedURI           string   `json:"feedUri"`       // [livemark]
	SiteURI           string   `json:"siteUri"`       // [livemark]

	HasDupe bool `json:"hasDupe"` // ??
}

type BookmarkCreatePayloadSchema struct {
	ID         string `json:"id"`         // [common]
	Type       string `json:"type"`       // [common]
	DateAdded  int64  `json:"dateAdded"`  // [common]
	ParentID   string `json:"parentid"`   // [common]
	ParentName string `json:"parentName"` // [common]

	Title             *string   `json:"title,omitempty"`         // [bookmark, microsummary, query, livemark, folder]
	URI               *string   `json:"bmkUri,omitempty"`        // [bookmark, microsummary, query]
	Description       *string   `json:"description,omitempty"`   // [bookmark, microsummary, query]
	LoadInSidebar     *bool     `json:"loadInSidebar,omitempty"` // [bookmark, microsummary, query]
	Tags              *[]string `json:"tags,omitempty"`          // [bookmark, microsummary, query]
	Keyword           *string   `json:"keyword,omitempty"`       // [bookmark, microsummary, query]
	Children          *[]string `json:"children,omitempty"`      // [folder, livemark]
	GeneratorUri      *string   `json:"generatorUri,omitempty"`  // [microsummary]
	StaticTitle       *string   `json:"staticTitle,omitempty"`   // [microsummary]
	FolderName        *string   `json:"folderName,omitempty"`    // [query]
	QueryID           *string   `json:"queryId,omitempty"`       // [query]
	SeparatorPosition *int      `json:"pos,omitempty"`           // [separator]
	FeedURI           *string   `json:"feedUri,omitempty"`       // [livemark]
	SiteURI           *string   `json:"siteUri,omitempty"`       // [livemark]
}

func (j BookmarkPayloadSchema) ToModel() BookmarkRecord {
	return BookmarkRecord{
		ID:                j.ID,
		Deleted:           j.Deleted,
		DateAdded:         optMilliTime(j.DateAdded),
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
		FeedURI:           j.FeedURI,
		SiteURI:           j.SiteURI,
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
	DateAdded         *time.Time
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
	FeedURI           string
	SiteURI           string
}

func (bm BookmarkRecord) ToJSON(ctx *cli.FFSContext) langext.H {
	r := langext.H{
		"id":         bm.ID,
		"type":       bm.Type,
		"deleted":    bm.Deleted,
		"added":      fmtOptDateToNullable(ctx, bm.DateAdded),
		"added_unix": fmtOptDateToNullableUnix(bm.DateAdded),
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
		r["feed-uri"] = bm.FeedURI
		r["site-uri"] = bm.SiteURI
		return r
	case BookmarkTypeSeparator:
		r["pos"] = bm.SeparatorPosition
		return r
	default:
		return r
	}
}

func (bm BookmarkRecord) ToSingleXML(ctx *cli.FFSContext, containsDeleted bool) any {

	switch bm.Type {
	case BookmarkTypeBookmark:
		type xmlentry struct {
			XMLName     xml.Name
			ID          string `xml:"ID,attr"`
			Deleted     string `xml:"Deleted,omitempty,attr"`
			Title       string `xml:"Title,attr"`
			URL         string `xml:"Href,attr"`
			AddDate     string `xml:"CDate,omitempty,attr"`
			Description string `xml:"Description,omitempty,attr"`
			Keyword     string `xml:"Keyword,omitempty,attr"`
			Tags        string `xml:"Tags,omitempty,attr"`
		}
		return xmlentry{
			XMLName:     xml.Name{Local: "Bookmark"},
			Deleted:     bm.formatDeleted(ctx, containsDeleted),
			ID:          bm.ID,
			Title:       bm.Title,
			URL:         bm.URI,
			AddDate:     fmtOptDate(ctx, bm.DateAdded),
			Description: bm.Description,
			Keyword:     bm.Keyword,
			Tags:        strings.Join(bm.Tags, ", "),
		}
	case BookmarkTypeMicroSummary:
		type xmlentry struct {
			XMLName      xml.Name
			ID           string `xml:"ID,attr"`
			Deleted      string `xml:"Deleted,omitempty,attr"`
			Title        string `xml:"Title,attr"`
			URL          string `xml:"Href,attr"`
			AddDate      string `xml:"CDate,omitempty,attr"`
			Description  string `xml:"Description,omitempty,attr"`
			Keyword      string `xml:"Keyword,omitempty,attr"`
			Tags         string `xml:"Tags,omitempty,attr"`
			GeneratorURI string `xml:"GeneratorURI,attr"`
			StaticTitle  string `xml:"StaticTitle,attr"`
		}
		return xmlentry{
			XMLName:      xml.Name{Local: "Microsummary"},
			Deleted:      bm.formatDeleted(ctx, containsDeleted),
			ID:           bm.ID,
			Title:        bm.Title,
			URL:          bm.URI,
			AddDate:      fmtOptDate(ctx, bm.DateAdded),
			Description:  bm.Description,
			Keyword:      bm.Keyword,
			Tags:         strings.Join(bm.Tags, ", "),
			GeneratorURI: bm.GeneratorUri,
			StaticTitle:  bm.StaticTitle,
		}
	case BookmarkTypeQuery:
		type xmlentry struct {
			XMLName     xml.Name
			ID          string `xml:"ID,attr"`
			Deleted     string `xml:"Deleted,omitempty,attr"`
			Title       string `xml:"Title,attr"`
			URL         string `xml:"Href,attr"`
			AddDate     string `xml:"Cdate,omitempty,attr"`
			Description string `xml:"Description,omitempty,attr"`
			Keyword     string `xml:"Keyword,omitempty,attr"`
			Tags        string `xml:"Tags,omitempty,attr"`
			FolderName  string `xml:"Foldername,attr"`
			QueryID     string `xml:"QueryID,attr"`
		}
		return xmlentry{
			XMLName:     xml.Name{Local: "Query"},
			Deleted:     bm.formatDeleted(ctx, containsDeleted),
			ID:          bm.ID,
			Title:       bm.Title,
			URL:         bm.URI,
			AddDate:     fmtOptDate(ctx, bm.DateAdded),
			Description: bm.Description,
			Keyword:     bm.Keyword,
			Tags:        strings.Join(bm.Tags, ", "),
			FolderName:  bm.FolderName,
			QueryID:     bm.QueryID,
		}
	case BookmarkTypeFolder:
		type xmlentry struct {
			XMLName  xml.Name
			ID       string `xml:"ID,attr"`
			Deleted  string `xml:"Deleted,omitempty,attr"`
			Title    string `xml:"Title,attr"`
			AddDate  string `xml:"CDate,omitempty,attr"`
			Children string `xml:"Children,attr"`
		}
		return xmlentry{
			XMLName:  xml.Name{Local: "Folder"},
			Deleted:  bm.formatDeleted(ctx, containsDeleted),
			ID:       bm.ID,
			Title:    bm.Title,
			AddDate:  fmtOptDate(ctx, bm.DateAdded),
			Children: strings.Join(bm.Children, ", "),
		}
	case BookmarkTypeLivemark:
		type xmlentry struct {
			XMLName  xml.Name
			ID       string `xml:"Id,attr"`
			Deleted  string `xml:"Deleted,omitempty,attr"`
			Title    string `xml:"Title,attr"`
			AddDate  string `xml:"CDate,omitempty,attr"`
			Children string `xml:"Children,attr"`
			FeedURI  string `xml:"FeedURI,attr"`
			SiteURI  string `xml:"SiteURI,attr"`
		}
		return xmlentry{
			XMLName:  xml.Name{Local: "Livemark"},
			Deleted:  bm.formatDeleted(ctx, containsDeleted),
			ID:       bm.ID,
			Title:    bm.Title,
			AddDate:  fmtOptDate(ctx, bm.DateAdded),
			Children: strings.Join(bm.Children, ", "),
			FeedURI:  bm.FeedURI,
			SiteURI:  bm.SiteURI,
		}
	case BookmarkTypeSeparator:
		type xmlentry struct {
			XMLName xml.Name
			ID      string `xml:"ID,attr"`
			Deleted string `xml:"Deleted,omitempty,attr"`
			AddDate string `xml:"CDate,omitempty,attr"`
			Pos     int    `xml:"Pos,attr"`
		}
		return xmlentry{
			XMLName: xml.Name{Local: "Separator"},
			Deleted: bm.formatDeleted(ctx, containsDeleted),
			ID:      bm.ID,
			Pos:     bm.SeparatorPosition,
		}
	default:
		type xmlentry struct {
			XMLName xml.Name
			ID      string `xml:"ID,attr"`
			Deleted string `xml:"Deleted,omitempty,attr"`
			AddDate string `xml:"CDate,attr"`
			Type    string `xml:"Type,attr"`
		}
		return xmlentry{
			XMLName: xml.Name{Local: "Unknown"},
			Type:    string(bm.Type),
			Deleted: bm.formatDeleted(ctx, containsDeleted),
			ID:      bm.ID,
		}
	}
}

func (bm BookmarkRecord) formatDeleted(ctx *cli.FFSContext, showFalse bool) string {
	if showFalse {
		return langext.FormatBool(bm.Deleted, "TRUE", "FALSE")
	} else {
		return langext.FormatBool(bm.Deleted, "TRUE", "")
	}
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

func (bmt BookmarkTreeRecord) ToTreeXML(ctx *cli.FFSContext, containsDeleted bool) any {
	if bmt.Type == BookmarkTypeFolder {
		arr := make([]any, 0, len(bmt.ResolvedChildren))
		for _, child := range bmt.ResolvedChildren {
			arr = append(arr, child.ToTreeXML(ctx, containsDeleted))
		}

		type xmlentry struct {
			XMLName  xml.Name
			Children []any
			ID       string `xml:"id,attr"`
			Deleted  string `xml:"deleted,omitempty,attr"`
			Title    string `xml:"title,attr"`
			AddDate  string `xml:"cdate,attr"`
		}
		return xmlentry{
			XMLName:  xml.Name{Local: "Folder"},
			Deleted:  bmt.formatDeleted(ctx, containsDeleted),
			ID:       bmt.ID,
			Title:    bmt.Title,
			AddDate:  fmtOptDate(ctx, bmt.DateAdded),
			Children: arr,
		}
	}
	if bmt.Type == BookmarkTypeLivemark {
		arr := make([]any, 0, len(bmt.ResolvedChildren))
		for _, child := range bmt.ResolvedChildren {
			arr = append(arr, child.ToTreeXML(ctx, containsDeleted))
		}

		type xmlentry struct {
			XMLName  xml.Name
			ID       string `xml:"id,attr"`
			Deleted  string `xml:"deleted,omitempty,attr"`
			Title    string `xml:"title,attr"`
			AddDate  string `xml:"cdate,attr"`
			Children []any
			FeedURI  string `xml:"feeduri,attr"`
			SiteURI  string `xml:"siteuri,attr"`
		}
		return xmlentry{
			XMLName:  xml.Name{Local: "Livemark"},
			Deleted:  bmt.formatDeleted(ctx, containsDeleted),
			ID:       bmt.ID,
			Title:    bmt.Title,
			AddDate:  fmtOptDate(ctx, bmt.DateAdded),
			Children: arr,
			FeedURI:  bmt.FeedURI,
			SiteURI:  bmt.SiteURI,
		}
	}

	return bmt.ToSingleXML(ctx, containsDeleted)
}
