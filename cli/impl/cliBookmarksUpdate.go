package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"ffsyncclient/langext"
	"ffsyncclient/models"
	"fmt"
	"github.com/joomcode/errorx"
	"strconv"
	"strings"
)

type CLIArgumentsBookmarksUpdate struct {
	RecordID string

	Title         *string
	URL           *string
	Description   *string
	LoadInSidebar *bool
	Tags          *[]string
	Keyword       *string
	Position      *int

	CLIArgumentsBookmarksUtil
}

func NewCLIArgumentsBookmarksUpdate() *CLIArgumentsBookmarksUpdate {
	return &CLIArgumentsBookmarksUpdate{
		Title:         nil,
		URL:           nil,
		Description:   nil,
		LoadInSidebar: nil,
		Tags:          nil,
		Keyword:       nil,
		Position:      nil,
	}
}

func (a *CLIArgumentsBookmarksUpdate) Mode() cli.Mode {
	return cli.ModeBookmarksUpdate
}

func (a *CLIArgumentsBookmarksUpdate) PositionArgCount() (*int, *int) {
	return langext.Ptr(1), langext.Ptr(1)
}

func (a *CLIArgumentsBookmarksUpdate) AvailableOutputFormats() []cli.OutputFormat {
	return []cli.OutputFormat{cli.OutputFormatText}
}

func (a *CLIArgumentsBookmarksUpdate) ShortHelp() [][]string {
	return [][]string{
		{"ffsclient bookmarks update <id>", "Partially update a bookmark"},
		{"          [--title <title>]", "Change the bookmark title"},
		{"          [--url <url>]", "Change the URL"},
		{"          [--description <desc>]", "Change the bookmark description"},
		{"          [--load-in-sidebar <true|false>]", "Set the `LoadInSidebar` field"},
		{"          [--tag <tag>]", "Change the tags, specify multiple times to set multiple tags"},
		{"          [--keyword <kw>]", "Specify the keyword (to activate the bookmark from the location bar)"},
		{"          [--position=<idx>]", "Change the position of the entry in the parent (0 = first). Can use negative indizes."},
	}
}

func (a *CLIArgumentsBookmarksUpdate) FullHelp() []string {
	return []string{
		"$> ffsclient bookmarks update <id> [--title <title>] [--url <url>] [--description <desc>] [--load-in-sidebar <true|false>] [--tag <tag>] [--keyword <kw>] [--position=<idx>]",
		"",
		"Update the specified fields of an existing bookmark entry.",
		"Supplied values that are not valid for the bookmark type result in an error.",
		"",
		"The fields of the found bookmark can be updated individually with the parameters:",
		"  * --title",
		"  * --url",
		"  * --description",
		"  * --load-in-sidebar",
		"  * --tag",
		"  * --keyword",
		"  * --position",
	}
}

func (a *CLIArgumentsBookmarksUpdate) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	a.RecordID = positionalArgs[0]

	for _, arg := range optionArgs {
		if arg.Key == "title" && arg.Value != nil {
			a.Title = langext.Ptr(*arg.Value)
			continue
		}
		if arg.Key == "url" && arg.Value != nil {
			a.URL = langext.Ptr(*arg.Value)
			continue
		}
		if arg.Key == "description" && arg.Value != nil {
			a.Description = langext.Ptr(*arg.Value)
			continue
		}
		if arg.Key == "load-in-sidebar" && arg.Value != nil {
			if strings.ToLower(*arg.Value) == "true" {
				a.LoadInSidebar = langext.Ptr(true)
			} else if strings.ToLower(*arg.Value) == "false" {
				a.LoadInSidebar = langext.Ptr(false)
			} else {
				return fferr.DirectOutput.New("Failed to parse boolean argument '--load-in-sidebar': '" + *arg.Value + "'")
			}
			continue
		}
		if arg.Key == "tag" && arg.Value != nil {
			if a.Tags == nil {
				a.Tags = langext.Ptr(make([]string, 0))
			}
			v := *a.Tags
			v = append(v, *arg.Value)
			a.Tags = &v
			continue
		}
		if arg.Key == "keyword" && arg.Value != nil {
			a.Keyword = langext.Ptr(*arg.Value)
			continue
		}
		if arg.Key == "position" && arg.Value != nil {
			if v, err := strconv.ParseInt(*arg.Value, 10, 32); err == nil {
				a.Position = langext.Ptr(int(v))
				continue
			}
			return fferr.DirectOutput.New("Failed to parse number argument '--position': '" + *arg.Value + "'")
		}
		return fferr.DirectOutput.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsBookmarksUpdate) Execute(ctx *cli.FFSContext) error {
	ctx.PrintVerbose("[Update Bookmark]")
	ctx.PrintVerbose("")
	ctx.PrintVerboseKV("RecordID", a.RecordID)

	// ========================================================================

	client, session, err := a.InitClient(ctx)
	if err != nil {
		return err
	}

	// ========================================================================

	ctx.PrintVerboseHeader("[0] Find bookmark")

	record, err := client.GetRecord(ctx, session, consts.CollectionBookmarks, a.RecordID, true)
	if err != nil && errorx.IsOfType(err, fferr.Request404) {
		return fferr.WrapDirectOutput(err, consts.ExitcodePasswordNotFound, "Record not found")
	}
	if err != nil {
		return errorx.Decorate(err, "failed to query record")
	}

	bmrec, err := models.UnmarshalBookmark(ctx, record)
	if err != nil {
		return errorx.Decorate(err, "failed to decode password-record")
	}

	// ========================================================================

	ctx.PrintVerboseHeader("[2] Patch Data")

	newData, err := a.patchData(ctx, record, bmrec)
	if err != nil {
		return err
	}

	// ========================================================================

	var parent models.BookmarkRecord
	var newParentPayload string

	if a.Position != nil {

		ctx.PrintVerboseHeader("[3] Query Parent")

		parentRecord, err := client.GetRecord(ctx, session, consts.CollectionBookmarks, bmrec.ParentID, true)
		if err != nil && errorx.IsOfType(err, fferr.Request404) {
			return errorx.Decorate(err, "failed to find parent").WithProperty(fferr.Exitcode, consts.ExitcodeRecordNotFound)
		}
		if err != nil {
			return errorx.Decorate(err, "failed to query parent-record")
		}

		bmparent, err := models.UnmarshalBookmark(ctx, parentRecord)
		if err != nil {
			return errorx.Decorate(err, "failed to decode parent-record")
		}

		bmrec, newPlainPayload, normpos, err := a.moveChild(ctx, parentRecord, bmparent, record.ID, *a.Position)
		if err != nil {
			return errorx.Decorate(err, "failed to calculate new pos in parent")
		}

		parent = bmrec
		newParentPayload = newPlainPayload

		if bmrec.Type == models.BookmarkTypeSeparator {

			ctx.PrintVerbose(fmt.Sprintf("Patch field [position] to %v", *a.Position))

			newData, err = langext.PatchJson(newData, "pos", normpos)
			if err != nil {
				return errorx.Decorate(err, "failed to patch data of existing record")
			}
		}
	}

	// ========================================================================

	if string(newData) != string(record.DecodedData) {

		ctx.PrintVerboseHeader("[4] Update record")

		newPayloadRecord, err := client.EncryptPayload(ctx, session, consts.CollectionBookmarks, string(newData))
		if err != nil {
			return err
		}

		update := models.RecordUpdate{
			ID:      a.RecordID,
			Payload: langext.Ptr(newPayloadRecord),
		}

		err = client.PutRecord(ctx, session, consts.CollectionBookmarks, update, false, false)
		if err != nil && errorx.IsOfType(err, fferr.Request404) {
			return fferr.WrapDirectOutput(err, consts.ExitcodeRecordNotFound, "Record not found")
		}
		if err != nil {
			return err
		}

	} else {

		ctx.PrintVerbose("Donot update record (nothing to do)")

	}

	// ========================================================================

	if a.Position != nil {

		ctx.PrintVerboseHeader("[5] Update parent")

		payloadParent, err := client.EncryptPayload(ctx, session, consts.CollectionBookmarks, newParentPayload)
		if err != nil {
			return err
		}

		updateParent := models.RecordUpdate{
			ID:      parent.ID,
			Payload: langext.Ptr(payloadParent),
		}

		err = client.PutRecord(ctx, session, consts.CollectionBookmarks, updateParent, false, false)
		if err != nil {
			return err
		}

	} else {

		ctx.PrintVerbose("Donot update parent (nothing to do)")

	}

	// ========================================================================

	if langext.Coalesce(ctx.Opt.Format, cli.OutputFormatText) != cli.OutputFormatText {
		return fferr.NewDirectOutput(consts.ExitcodeUnsupportedOutputFormat, "Unsupported output-format: "+ctx.Opt.Format.String())
	}

	ctx.PrintPrimaryOutput("Okay.")
	return nil
}

func (a *CLIArgumentsBookmarksUpdate) patchData(ctx *cli.FFSContext, record models.Record, bmrec models.BookmarkRecord) ([]byte, error) {
	var err error

	newData := record.DecodedData

	if a.Title != nil {
		if !langext.InArray(bmrec.Type, []models.BookmarkType{models.BookmarkTypeBookmark, models.BookmarkTypeFolder}) {
			return nil, fferr.NewDirectOutput(consts.ExitcodeBookmarkFieldNotSupported, fmt.Sprintf("The field 'tile' is not supported on bookmarks of the type %s", bmrec.Type))
		}

		ctx.PrintVerbose(fmt.Sprintf("Patch field [title] from \"%s\" to \"%s\"", bmrec.Title, *a.Title))

		newData, err = langext.PatchJson(newData, "title", *a.Title)
		if err != nil {
			return nil, errorx.Decorate(err, "failed to patch data of existing record")
		}
	}

	if a.URL != nil {
		if !langext.InArray(bmrec.Type, []models.BookmarkType{models.BookmarkTypeBookmark}) {
			return nil, fferr.NewDirectOutput(consts.ExitcodeBookmarkFieldNotSupported, fmt.Sprintf("The field 'tile' is not supported on bookmarks of the type %s", bmrec.Type))
		}

		ctx.PrintVerbose(fmt.Sprintf("Patch field [url] from \"%s\" to \"%s\"", bmrec.URI, *a.URL))

		newData, err = langext.PatchJson(newData, "bmkUri", *a.URL)
		if err != nil {
			return nil, errorx.Decorate(err, "failed to patch data of existing record")
		}
	}

	if a.Description != nil {
		if !langext.InArray(bmrec.Type, []models.BookmarkType{models.BookmarkTypeBookmark}) {
			return nil, fferr.NewDirectOutput(consts.ExitcodeBookmarkFieldNotSupported, fmt.Sprintf("The field 'tile' is not supported on bookmarks of the type %s", bmrec.Type))
		}

		ctx.PrintVerbose(fmt.Sprintf("Patch field [description] from \"%s\" to \"%s\"", bmrec.Description, *a.Description))

		newData, err = langext.PatchJson(newData, "description", *a.Description)
		if err != nil {
			return nil, errorx.Decorate(err, "failed to patch data of existing record")
		}
	}

	if a.LoadInSidebar != nil {
		if !langext.InArray(bmrec.Type, []models.BookmarkType{models.BookmarkTypeBookmark}) {
			return nil, fferr.NewDirectOutput(consts.ExitcodeBookmarkFieldNotSupported, fmt.Sprintf("The field 'tile' is not supported on bookmarks of the type %s", bmrec.Type))
		}

		ctx.PrintVerbose(fmt.Sprintf("Patch field [loadInSidebar] from \"%v\" to \"%v\"", bmrec.LoadInSidebar, *a.LoadInSidebar))

		newData, err = langext.PatchJson(newData, "loadInSidebar", *a.LoadInSidebar)
		if err != nil {
			return nil, errorx.Decorate(err, "failed to patch data of existing record")
		}
	}

	if a.Tags != nil {
		if !langext.InArray(bmrec.Type, []models.BookmarkType{models.BookmarkTypeBookmark}) {
			return nil, fferr.NewDirectOutput(consts.ExitcodeBookmarkFieldNotSupported, fmt.Sprintf("The field 'tile' is not supported on bookmarks of the type %s", bmrec.Type))
		}

		ctx.PrintVerbose(fmt.Sprintf("Patch field [tags] from [%v] to [%v]", strings.Join(bmrec.Tags, ", "), strings.Join(*a.Tags, ", ")))

		newData, err = langext.PatchJson(newData, "tags", *a.Tags)
		if err != nil {
			return nil, errorx.Decorate(err, "failed to patch data of existing record")
		}
	}

	if a.Keyword != nil {
		if !langext.InArray(bmrec.Type, []models.BookmarkType{models.BookmarkTypeBookmark}) {
			return nil, fferr.NewDirectOutput(consts.ExitcodeBookmarkFieldNotSupported, fmt.Sprintf("The field 'tile' is not supported on bookmarks of the type %s", bmrec.Type))
		}

		ctx.PrintVerbose(fmt.Sprintf("Patch field [keyword] from \"%v\" to \"%v\"", bmrec.Keyword, *a.Keyword))

		newData, err = langext.PatchJson(newData, "keyword", *a.Keyword)
		if err != nil {
			return nil, errorx.Decorate(err, "failed to patch data of existing record")
		}
	}
	return newData, nil
}
