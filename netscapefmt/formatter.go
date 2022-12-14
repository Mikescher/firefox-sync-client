package netscapefmt

import (
	"bytes"
	"encoding/xml"
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/models"
	"fmt"
	"strings"
)

// https://docs.microsoft.com/en-us/previous-versions/windows/internet-explorer/ie-developer/platform-apis/aa753582(v=vs.85)

func Format(ctx *cli.FFSContext, records []*models.BookmarkTreeRecord) string {
	printer := &ncPrinter{}

	printer.appendLine("<!DOCTYPE NETSCAPE-Bookmark-file-1>")
	printer.appendLine("<!-- This file was generated by firefox-sync-client -->")
	printer.appendLine("<META HTTP-EQUIV=\"Content-Type\" CONTENT=\"text/html; charset=UTF-8\">")
	printer.appendLine("<meta http-equiv=\"Content-Security-Policy\"")
	printer.appendLine("      content=\"default-src 'self'; script-src 'none'; img-src data: *; object-src 'none'\"></meta>")
	printer.appendLine("<TITLE>Bookmarks</TITLE>")
	printer.appendLine("<H1>Bookmarks Menu</H1>")
	printer.appendLine("")

	printer.appendLine("<DL><p>")
	printer.inc()
	for _, v := range records {
		if v.ID != consts.BookmarkIDToolbar {
			for _, child := range v.ResolvedChildren {
				printItem(ctx, printer, child)
			}
		}
	}
	for _, v := range records {
		if v.ID == consts.BookmarkIDToolbar {
			itemstr := "<DT><H3"
			if v.DateAdded != nil {
				itemstr += fmt.Sprintf(" ADD_DATE=\"%d\"", v.DateAdded.Unix())
				itemstr += fmt.Sprintf(" LAST_MODIFIED=\"%d\"", v.DateAdded.Unix())
			}
			itemstr += " PERSONAL_TOOLBAR_FOLDER=\"true\""
			itemstr += fmt.Sprintf(">%s</H3>", escape("Bookmarks Toolbar"))
			printer.appendLine(itemstr)
			printer.appendLine("<DL><p>")
			printer.inc()
			for _, child := range v.ResolvedChildren {
				printItem(ctx, printer, child)
			}
			printer.dec()
			printer.appendLine("</DL><p>")
		}
	}
	printer.dec()
	printer.appendLine("</DL>")

	return printer.String()
}

func printItem(ctx *cli.FFSContext, printer *ncPrinter, item *models.BookmarkTreeRecord) {
	switch item.Type {
	case models.BookmarkTypeBookmark:
		itemstr := "<DT><A"
		itemstr += fmt.Sprintf(" HREF=\"%s\"", escape(item.URI))
		if item.DateAdded != nil {
			itemstr += fmt.Sprintf(" ADD_DATE=\"%d\"", item.DateAdded.Unix())
			itemstr += fmt.Sprintf(" LAST_MODIFIED=\"%d\"", item.DateAdded.Unix())
		}
		if item.Keyword != "" {
			itemstr += fmt.Sprintf(" SHORTCUTURL=\"%s\"", escape(item.Keyword))
		}
		if len(item.Tags) > 0 {
			itemstr += fmt.Sprintf(" TAGS=\"%s\"", escape(strings.Join(item.Tags, ",")))
		}
		itemstr += fmt.Sprintf(">%s</A>", escape(item.Title))

		printer.appendLine(itemstr)
		return

	case models.BookmarkTypeFolder:
		itemstr := "<DT><H3"
		if item.DateAdded != nil {
			itemstr += fmt.Sprintf(" ADD_DATE=\"%d\"", item.DateAdded.Unix())
			itemstr += fmt.Sprintf(" LAST_MODIFIED=\"%d\"", item.DateAdded.Unix())
		}
		itemstr += fmt.Sprintf(">%s</H3>", escape(item.Title))

		printer.appendLine(itemstr)

		printer.appendLine("<DL><p>")
		printer.inc()
		for _, child := range item.ResolvedChildren {
			printItem(ctx, printer, child)
		}
		printer.dec()
		printer.appendLine("</DL><p>")
		return

	case models.BookmarkTypeSeparator:
		printer.appendLine("<HR>")
		return

	default:
		ctx.PrintVerbose(fmt.Sprintf("[WARN] skip item of type %v in netscape-formatter", item.Type))
	}
}

func escape(v string) string {
	var buffy bytes.Buffer
	xml.Escape(&buffy, []byte(v))
	return buffy.String()
}
