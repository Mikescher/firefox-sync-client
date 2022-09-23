package impl

import (
	"encoding/json"
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"ffsyncclient/langext"
	"ffsyncclient/models"
	"ffsyncclient/syncclient"
	"github.com/joomcode/errorx"
	"strconv"
	"time"
)

type CLIArgumentsBookmarksCreateFolder struct {
	Title    string
	ParentID string
	Position int

	CLIArgumentsBookmarksUtil
}

func NewCLIArgumentsBookmarksCreateFolder() *CLIArgumentsBookmarksCreateFolder {
	return &CLIArgumentsBookmarksCreateFolder{
		ParentID:                  "unfiled",
		Position:                  -1,
		CLIArgumentsBookmarksUtil: CLIArgumentsBookmarksUtil{},
	}
}

func (a *CLIArgumentsBookmarksCreateFolder) Mode() cli.Mode {
	return cli.ModeBookmarksCreateFolder
}

func (a *CLIArgumentsBookmarksCreateFolder) PositionArgCount() (*int, *int) {
	return langext.Ptr(1), langext.Ptr(1)
}

func (a *CLIArgumentsBookmarksCreateFolder) ShortHelp() [][]string {
	return [][]string{
		{"ffsclient bookmarks create folder <title>", "Insert a new bookmark"},
		{"          [--parent <id>]", "Specify the ID of the parent folder (if not specified the entry lives under `unfiled`)"},
		{"          [--position=<idx>]", "The position of the entry in the parent (0 = first, default is last). Can use negative indizes."},
	}
}

func (a *CLIArgumentsBookmarksCreateFolder) FullHelp() []string {
	return []string{
		"$> ffsclient bookmarks create folder <title> [--parent <id>] [--position <idx>]",
		"",
		"Create a new bookmark with the type [folder]",
		"",
		"The field <title> must be specified.",
		"With --parent you can specify the ID of the parent folder. Throws an error if the parent does not exist or is not an folder. The default value is `unfiled`",
		"With --position you can specify the position in the parent folder. The left-most position is 0 and the last position is len(folder.children). You can also use negative indizes: -1 is the last position and -2 the second-last etc. An invalid position throws an error.",
		"If the position is negative you _have_ to use the --position=XX syntax. (Writing `--position XX` will result in a parser error)",
		"",
		"Outputs the RecordID of the newly created entry on success.",
	}
}

func (a *CLIArgumentsBookmarksCreateFolder) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	a.Title = positionalArgs[0]

	for _, arg := range optionArgs {
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

func (a *CLIArgumentsBookmarksCreateFolder) Execute(ctx *cli.FFSContext) int {
	ctx.PrintVerbose("[Create Bookmark<Folder>]")
	ctx.PrintVerbose("")

	// ========================================================================

	cfp, err := ctx.AbsSessionFilePath()
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	if !langext.FileExists(cfp) {
		ctx.PrintFatalMessage("Sessionfile does not exist.")
		ctx.PrintFatalMessage("Use `ffsclient login <email> <password>` first")
		return consts.ExitcodeNoLogin
	}

	// ========================================================================

	client := syncclient.NewFxAClient(ctx.Opt.AuthServerURL)

	ctx.PrintVerbose("Load existing session from " + cfp)
	session, err := syncclient.LoadSession(ctx, cfp)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	session, err = client.AutoRefreshSession(ctx, session)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	// ========================================================================

	recordID := a.newBookmarkID()

	now := time.Now()

	ctx.PrintVerboseHeader("[1] Search for parent")

	parent, newParentPayload, err, excode := a.calculateParent(ctx, client, session, recordID, a.ParentID, a.Position)
	if err != nil {
		ctx.PrintFatalError(errorx.Decorate(err, "failed to find+calculate parent"))
		return excode
	}

	ctx.PrintVerbose("Found Record Parent record: '" + parent.ID + "'")

	ctx.PrintVerboseHeader("[2] Create new record")

	bso := models.BookmarkCreatePayloadSchema{
		ID:         recordID,
		Type:       string(models.BookmarkTypeFolder),
		DateAdded:  now.UnixMilli(),
		ParentID:   parent.ID,
		ParentName: parent.Title,

		Title: langext.Ptr(a.Title),
	}

	plainPayload, err := json.Marshal(bso)
	if err != nil {
		ctx.PrintFatalError(errorx.Decorate(err, "failed to marshal BSO json"))
		return consts.ExitcodeError
	}

	payloadNewRecord, err := client.EncryptPayload(ctx, session, consts.CollectionBookmarks, string(plainPayload))
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	update := models.RecordUpdate{
		ID:      recordID,
		Payload: langext.Ptr(payloadNewRecord),
	}

	err = client.PutRecord(ctx, session, consts.CollectionBookmarks, update, true, false)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	ctx.PrintVerboseHeader("[3] Update parent record")

	payloadParent, err := client.EncryptPayload(ctx, session, consts.CollectionBookmarks, newParentPayload)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	updateParent := models.RecordUpdate{
		ID:      parent.ID,
		Payload: langext.Ptr(payloadParent),
	}

	err = client.PutRecord(ctx, session, consts.CollectionBookmarks, updateParent, false, false)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	if langext.Coalesce(ctx.Opt.Format, cli.OutputFormatText) != cli.OutputFormatText {
		ctx.PrintFatalMessage("Unsupported output-format: " + ctx.Opt.Format.String())
		return consts.ExitcodeUnsupportedOutputFormat
	}

	ctx.PrintPrimaryOutput(recordID)
	return 0
}
