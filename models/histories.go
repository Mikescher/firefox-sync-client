package models

import (
	"encoding/xml"
	"ffsyncclient/cli"
	"fmt"
	"git.blackforestbytes.com/BlackForestBytes/goext/langext"
	"time"
)

type HistoryPayloadSchema struct {
	ID      string               `json:"id"`
	Deleted bool                 `json:"deleted,omitempty"`
	URI     string               `json:"histUri"`
	Title   string               `json:"title"`
	Visits  []HistoryVisitSchema `json:"visits"`
}

type HistoryVisitSchema struct {
	TransitionType HistoryTransitionType `json:"type"`
	VisitDate      int64                 `json:"date"` // (!) This is microseconds, not milliseconds like other dates
}

type HistoryTransitionType int

const (
	HistoryTransitionTypeLink              HistoryTransitionType = 1 // This transition type means the user followed a link and got a new toplevel window.
	HistoryTransitionTypeTyped             HistoryTransitionType = 2 // This transition type is set when the user typed the URL to get to the page.
	HistoryTransitionTypeBookmark          HistoryTransitionType = 3 // This transition type is set when the user followed a bookmark to get to the page.
	HistoryTransitionTypeEmbed             HistoryTransitionType = 4 // This transition type is set when some inner content is loaded.
	HistoryTransitionTypeRedirectPermanent HistoryTransitionType = 5 // This transition type is set when the transition was a permanent redirect.
	HistoryTransitionTypeRedirectTemporary HistoryTransitionType = 6 // This transition type is set when the transition was a temporary redirect.
	HistoryTransitionTypeDownload          HistoryTransitionType = 7 // This transition type is set when the transition is a download.
	HistoryTransitionTypeFramedLink        HistoryTransitionType = 8 // This transition type is set when the user followed a link that loaded a page in a frame.
	HistoryTransitionTypeReload            HistoryTransitionType = 9 // This transition type means the page has been reloaded.
)

func (t HistoryTransitionType) ConstantString() string {
	switch t {
	case HistoryTransitionTypeLink:
		return "LINK"
	case HistoryTransitionTypeTyped:
		return "TYPED"
	case HistoryTransitionTypeBookmark:
		return "BOOKMARK"
	case HistoryTransitionTypeEmbed:
		return "EMBED"
	case HistoryTransitionTypeRedirectPermanent:
		return "REDIRECT_PERMANENT"
	case HistoryTransitionTypeRedirectTemporary:
		return "REDIRECT_TEMPORARY"
	case HistoryTransitionTypeDownload:
		return "DOWNLOAD"
	case HistoryTransitionTypeFramedLink:
		return "FRAMED_LINK"
	case HistoryTransitionTypeReload:
		return "RELOAD"
	default:
		return fmt.Sprintf("UNKNOWN_%d", int(t))
	}
}

func (j HistoryPayloadSchema) ToModel() HistoryRecord {
	visits := make([]HistoryVisit, 0, len(j.Visits))
	for _, v := range j.Visits {
		visits = append(visits, HistoryVisit{
			TransitionType: v.TransitionType,
			VisitDate:      time.UnixMicro(v.VisitDate), // UnixMicro, not UnixMilli (!)
		})
	}

	return HistoryRecord{
		ID:      j.ID,
		Deleted: j.Deleted,
		URI:     j.URI,
		Title:   j.Title,
		Visits:  visits,
	}
}

type HistoryRecord struct {
	ID      string
	Deleted bool
	URI     string
	Title   string
	Visits  []HistoryVisit
}

func (r HistoryRecord) ToJSON(ctx *cli.FFSContext) langext.H {
	visits := make([]langext.H, 0, len(r.Visits))
	for _, v := range r.Visits {
		visits = append(visits, langext.H{
			"date":             v.VisitDate.In(ctx.Opt.TimeZone).Format(ctx.Opt.TimeFormat),
			"date_unix":        v.VisitDate.Unix(),
			"transition":       v.TransitionType.ConstantString(),
			"transition_const": int(v.TransitionType),
		})
	}
	return langext.H{
		"id":      r.ID,
		"deleted": r.Deleted,
		"uri":     r.URI,
		"title":   r.Title,
		"visits":  visits,
	}
}

func (r HistoryRecord) ToSingleXML(ctx *cli.FFSContext, containsDeleted bool) any {
	type visitentry struct {
		XMLName    xml.Name
		Date       string `xml:"date,attr"`
		Transition string `xml:"transition,attr"`
	}
	type xmlentry struct {
		XMLName xml.Name
		ID      string `xml:"ID,attr"`
		Deleted string `xml:"Deleted,omitempty,attr"`
		Uri     string `xml:"URI,attr"`
		Title   string `xml:"Title,attr"`
		Visits  []visitentry
	}
	visits := make([]visitentry, 0, len(r.Visits))
	for _, v := range r.Visits {
		visits = append(visits, visitentry{
			XMLName:    xml.Name{Local: "Visit"},
			Date:       v.VisitDate.In(ctx.Opt.TimeZone).Format(ctx.Opt.TimeFormat),
			Transition: v.TransitionType.ConstantString(),
		})
	}
	return xmlentry{
		XMLName: xml.Name{Local: "Entry"},
		ID:      r.ID,
		Deleted: r.formatDeleted(ctx, containsDeleted),
		Uri:     r.URI,
		Title:   r.Title,
		Visits:  visits,
	}
}

func (r HistoryRecord) formatDeleted(ctx *cli.FFSContext, showFalse bool) string {
	if showFalse {
		return langext.FormatBool(r.Deleted, "TRUE", "FALSE")
	} else {
		return langext.FormatBool(r.Deleted, "TRUE", "")
	}
}

func (r HistoryRecord) LastVisitStr(ctx *cli.FFSContext) string {
	if len(r.Visits) == 0 {
		return ""
	}
	return r.Visits[len(r.Visits)-1].VisitDate.In(ctx.Opt.TimeZone).Format(ctx.Opt.TimeFormat)
}

func (r HistoryRecord) FirstVisitStr(ctx *cli.FFSContext) string {
	if len(r.Visits) == 0 {
		return ""
	}
	return r.Visits[0].VisitDate.In(ctx.Opt.TimeZone).Format(ctx.Opt.TimeFormat)
}

type HistoryVisit struct {
	TransitionType HistoryTransitionType
	VisitDate      time.Time
}
