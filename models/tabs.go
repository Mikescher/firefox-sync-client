package models

import (
	"encoding/xml"
	"ffsyncclient/cli"
	"fmt"
	"git.blackforestbytes.com/BlackForestBytes/goext/langext"
	"time"
)

type TabPayloadSchema struct {
	ID         string                   `json:"id"`
	Deleted    bool                     `json:"deleted,omitempty"`
	ClientName string                   `json:"clientName"`
	Tabs       []SingleTabPayloadSchema `json:"tabs"`
}

type SingleTabPayloadSchema struct {
	Title      string   `json:"title"`
	UrlHistory []string `json:"urlHistory"`
	Icon       string   `json:"icon"`
	LastUsed   int64    `json:"lastUsed"`
}

func (j TabPayloadSchema) ToMultiModels() []TabRecord {
	return langext.ArrMapExt(j.Tabs, func(idx int, v SingleTabPayloadSchema) TabRecord {
		return TabRecord{
			ClientID:      j.ID,
			ClientDeleted: j.Deleted,
			ClientName:    j.ClientName,
			Index:         idx,
			Title:         v.Title,
			UrlHistory:    langext.ForceArray(v.UrlHistory),
			Icon:          v.Icon,
			LastUsed:      time.Unix(v.LastUsed, 0), // UnixSeconds, not UnixMilli (!)
		}
	})
}

func (j TabPayloadSchema) ToClientModel() TabClientRecord {
	return TabClientRecord{
		ID:      j.ID,
		Deleted: j.Deleted,
		Name:    j.ClientName,
		Tabs:    j.ToMultiModels(),
	}
}

type TabRecord struct {
	ClientID      string
	ClientDeleted bool
	ClientName    string
	Index         int
	Title         string
	UrlHistory    []string
	Icon          string
	LastUsed      time.Time
}

type TabClientRecord struct {
	ID      string
	Deleted bool
	Name    string
	Tabs    []TabRecord
}

func (bm TabRecord) ToJSON(ctx *cli.FFSContext) langext.H {
	return langext.H{
		"client_id":      bm.ClientID,
		"client_deleted": bm.ClientDeleted,
		"client_name":    bm.ClientName,
		"index":          bm.Index,
		"title":          bm.Title,
		"urlHistory":     bm.UrlHistory,
		"icon":           bm.Icon,
		"lastUsed":       bm.LastUsed.In(ctx.Opt.TimeZone).Format(ctx.Opt.TimeFormat),
		"lastUsed_unix":  bm.LastUsed.Unix(),
	}
}

func (bm TabClientRecord) ToSingleXML(ctx *cli.FFSContext, containsDeleted bool) any {
	type xmlentry struct {
		XMLName xml.Name

		ID      string `xml:"ID,attr"`
		Deleted string `xml:"Deleted,omitempty,attr"`
		Name    string `xml:"Name,attr"`

		Tabs []any `xml:"Tab"`
	}
	return xmlentry{
		XMLName: xml.Name{Local: "Tab"},
		Deleted: bm.formatDeleted(ctx, containsDeleted),
		ID:      bm.ID,
		Name:    bm.Name,
		Tabs:    langext.ArrMap(bm.Tabs, func(v TabRecord) any { return v.toSingleXML(ctx) }),
	}
}

func (bm TabRecord) toSingleXML(ctx *cli.FFSContext) any {
	type histentry struct {
		XMLName xml.Name

		Value string `xml:",chardata"`
	}
	type xmlentry struct {
		XMLName xml.Name

		Index        string `xml:"Index,attr"`
		Title        string `xml:"Title,attr"`
		Icon         string `xml:"Icon,attr"`
		LastUsed     string `xml:"LastUsed,attr"`
		LastUsedUnix int64  `xml:"LastUsedUnix,attr"`

		History []histentry `xml:"history"`
	}
	return xmlentry{
		XMLName:      xml.Name{Local: "Tab"},
		Index:        fmt.Sprintf("%d", bm.Index),
		Title:        bm.Title,
		Icon:         bm.Icon,
		LastUsed:     bm.LastUsed.In(ctx.Opt.TimeZone).Format(ctx.Opt.TimeFormat),
		LastUsedUnix: bm.LastUsed.Unix(),
		History:      langext.ArrMap(bm.UrlHistory, func(v string) histentry { return histentry{XMLName: xml.Name{Local: "history"}, Value: v} }),
	}
}

func (bm TabClientRecord) formatDeleted(ctx *cli.FFSContext, showFalse bool) string {
	if showFalse {
		return langext.FormatBool(bm.Deleted, "TRUE", "FALSE")
	} else {
		return langext.FormatBool(bm.Deleted, "TRUE", "")
	}
}
