package impl

import (
	"encoding/json"
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"ffsyncclient/langext"
	"ffsyncclient/models"
	"github.com/joomcode/errorx"
	"strconv"
	"time"
)

type CLIArgumentsBookmarksCreateBookmark struct {
	Title         string
	URL           string
	Description   string
	LoadInSidebar bool
	Tags          []string
	Keyword       string
	ParentID      string
	Position      int

	CLIArgumentsBookmarksUtil
}

func NewCLIArgumentsBookmarksCreateBookmark() *CLIArgumentsBookmarksCreateBookmark {
	return &CLIArgumentsBookmarksCreateBookmark{
		Description:   "",
		LoadInSidebar: false,
		Tags:          make([]string, 0),
		Keyword:       "",
		ParentID:      "unfiled",
		Position:      -1,
	}
}

func (a *CLIArgumentsBookmarksCreateBookmark) Mode() cli.Mode {
	return cli.ModeBookmarksCreateBookmark
}

func (a *CLIArgumentsBookmarksCreateBookmark) PositionArgCount() (*int, *int) {
	return langext.Ptr(2), langext.Ptr(2)
}

func (a *CLIArgumentsBookmarksCreateBookmark) AvailableOutputFormats() []cli.OutputFormat {
	return []cli.OutputFormat{cli.OutputFormatText}
}

func (a *CLIArgumentsBookmarksCreateBookmark) ShortHelp() [][]string {
	return [][]string{
		{"ffsclient bookmarks create bookmark <title> <url>", "Insert a new bookmark"},
		{"          [--description <desc>]", "Specify the bookmark description"},
		{"          [--load-in-sidebar]", "If specified the `LoadInSidebar` field is set to true (default is false)"},
		{"          [--tag <tag>]", "Add a tag to the bookmark, specify multiple times to add multiple tags"},
		{"          [--keyword <kw>]", "Specify the keyword (to activate the bookmark from the location bar)"},
		{"          [--parent <id>]", "Specify the ID of the parent folder (if not specified the entry lives under `unfiled`)"},
		{"          [--position=<idx>]", "The position of the entry in the parent (0 = first, default is last). Can use negative indizes."},
	}
}

func (a *CLIArgumentsBookmarksCreateBookmark) FullHelp() []string {
	return []string{
		"$> ffsclient bookmarks create bookmark <title> <url> [--description <desc>] [--load-in-sidebar] [--tag <tag>] [--keyword <kw>] [--parent <id>] [--position <idx>]",
		"",
		"Create a new bookmark with the type [bookmark]",
		"",
		"The fields <title> and <url> must be specified.",
		"If --load-in-sidebar is not specified the default value of false is used.",
		"You can specify one or more tags by supplying multiple --tag parameter.",
		"With --keyword you can specify an alias to activate the bookmark from the location bar.",
		"With --parent you can specify the ID of the parent folder. Throws an error if the parent does not exist or is not an folder. The default value is `unfiled`",
		"With --position you can specify the position in the parent folder. The left-most position is 0 and the last position is len(folder.children). You can also use negative indizes: -1 is the last position and -2 the second-last etc. An invalid position throws an error.",
		"If the position is negative you _have_ to use the --position=XX syntax. (Writing `--position XX` will result in a parser error)",
		"",
		"Outputs the RecordID of the newly created entry on success.",
	}
}

func (a *CLIArgumentsBookmarksCreateBookmark) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	a.Title = positionalArgs[0]
	a.URL = positionalArgs[1]

	for _, arg := range optionArgs {
		if arg.Key == "description" && arg.Value != nil {
			a.Description = *arg.Value
			continue
		}
		if arg.Key == "load-in-sidebar" && arg.Value == nil {
			a.LoadInSidebar = true
			continue
		}
		if arg.Key == "tag" && arg.Value != nil {
			a.Tags = append(a.Tags, *arg.Value)
			continue
		}
		if arg.Key == "keyword" && arg.Value != nil {
			a.Keyword = *arg.Value
			continue
		}
		if arg.Key == "parent" && arg.Value != nil {
			a.ParentID = *arg.Value
			continue
		}
		if arg.Key == "position" && arg.Value != nil {
			if v, err := strconv.ParseInt(*arg.Value, 10, 32); err == nil {
				a.Position = int(v)
				continue
			}
			return fferr.DirectOutput.New("Failed to parse number argument '--position': '" + *arg.Value + "'")
		}
		return fferr.DirectOutput.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsBookmarksCreateBookmark) Execute(ctx *cli.FFSContext) error {
	ctx.PrintVerbose("[Create Bookmark<Bookmark>]")
	ctx.PrintVerbose("")

	// ========================================================================

	client, session, err := a.InitClient(ctx)
	if err != nil {
		return err
	}

	// ========================================================================

	recordID := a.newBookmarkID()

	now := time.Now()

	ctx.PrintVerboseHeader("[1] Search for parent")

	parent, newParentPayload, _, err := a.calculateParent(ctx, client, session, recordID, a.ParentID, a.Position)
	if err != nil {
		return errorx.Decorate(err, "failed to find+calculate parent")
	}

	ctx.PrintVerbose("Found Record Parent record: '" + parent.ID + "'")

	ctx.PrintVerboseHeader("[2] Create new record")

	bso := models.BookmarkCreatePayloadSchema{
		ID:         recordID,
		Type:       string(models.BookmarkTypeBookmark),
		DateAdded:  now.UnixMilli(),
		ParentID:   parent.ID,
		ParentName: parent.Title,

		Title:         langext.Ptr(a.Title),
		URI:           langext.Ptr(a.URL),
		Description:   langext.Ptr(a.Description),
		LoadInSidebar: langext.Ptr(a.LoadInSidebar),
		Tags:          langext.Ptr(a.Tags),
		Keyword:       langext.Ptr(a.Keyword),
	}

	plainPayload, err := json.Marshal(bso)
	if err != nil {
		return errorx.Decorate(err, "failed to marshal BSO json")
	}

	payloadNewRecord, err := client.EncryptPayload(ctx, session, consts.CollectionBookmarks, string(plainPayload))
	if err != nil {
		return err
	}

	update := models.RecordUpdate{
		ID:      recordID,
		Payload: langext.Ptr(payloadNewRecord),
	}

	err = client.PutRecord(ctx, session, consts.CollectionBookmarks, update, true, false)
	if err != nil {
		return err
	}

	ctx.PrintVerboseHeader("[3] Update parent record")

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

	// ========================================================================

	if langext.Coalesce(ctx.Opt.Format, cli.OutputFormatText) != cli.OutputFormatText {
		return fferr.NewDirectOutput(consts.ExitcodeUnsupportedOutputFormat, "Unsupported output-format: "+ctx.Opt.Format.String())
	}

	ctx.PrintPrimaryOutput(recordID)
	return nil
}
