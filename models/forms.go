package models

import (
	"encoding/xml"
	"ffsyncclient/cli"
	"gogs.mikescher.com/BlackForestBytes/goext/langext"
	"time"
)

type FormPayloadSchema struct {
	ID      string `json:"id"`
	Deleted bool   `json:"deleted,omitempty"`
	Name    string `json:"name"`
	Value   string `json:"value"`
}

type FormCreatePayloadSchema struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (j FormPayloadSchema) ToModel(r Record) FormRecord {
	return FormRecord{
		ID:           j.ID,
		LastModified: r.Modified,
		Deleted:      j.Deleted,
		Name:         j.Name,
		Value:        j.Value,
	}
}

type FormRecord struct {
	ID           string
	LastModified time.Time
	Deleted      bool
	Name         string
	Value        string
}

func (bm FormRecord) ToJSON(ctx *cli.FFSContext) langext.H {
	return langext.H{
		"id":                bm.ID,
		"deleted":           bm.Deleted,
		"name":              bm.Name,
		"value":             bm.Value,
		"lastModified":      bm.LastModified.Format(ctx.Opt.TimeFormat),
		"lastModified_unix": bm.LastModified.Unix(),
	}
}

func (bm FormRecord) ToSingleXML(ctx *cli.FFSContext, containsDeleted bool) any {
	type xmlentry struct {
		XMLName xml.Name
		ID      string `xml:"id,attr"`
		Deleted string `xml:"deleted,omitempty,attr"`
		Name    string `xml:"name,attr"`
		Value   string `xml:"value,attr"`
		Date    string `xml:"mdate,omitempty,attr"`
	}
	return xmlentry{
		XMLName: xml.Name{Local: "form"},
		Deleted: bm.formatDeleted(ctx, containsDeleted),
		ID:      bm.ID,
		Name:    bm.Name,
		Value:   bm.Value,
		Date:    bm.LastModified.Format(ctx.Opt.TimeFormat),
	}
}

func (bm FormRecord) formatDeleted(ctx *cli.FFSContext, showFalse bool) string {
	if showFalse {
		return langext.FormatBool(bm.Deleted, "TRUE", "FALSE")
	} else {
		return langext.FormatBool(bm.Deleted, "TRUE", "")
	}
}
