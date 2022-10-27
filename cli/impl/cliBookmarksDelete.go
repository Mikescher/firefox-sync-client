package impl

import (
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"ffsyncclient/models"
	"fmt"
	"github.com/joomcode/errorx"
	"gogs.mikescher.com/BlackForestBytes/goext/langext"
)

type CLIArgumentsBookmarksDelete struct {
	RecordID string

	CLIArgumentsBookmarksUtil
}

func NewCLIArgumentsBookmarksDelete() *CLIArgumentsBookmarksDelete {
	return &CLIArgumentsBookmarksDelete{}
}

func (a *CLIArgumentsBookmarksDelete) Mode() cli.Mode {
	return cli.ModeBookmarksDelete
}

func (a *CLIArgumentsBookmarksDelete) PositionArgCount() (*int, *int) {
	return langext.Ptr(1), langext.Ptr(1)
}

func (a *CLIArgumentsBookmarksDelete) AvailableOutputFormats() []cli.OutputFormat {
	return []cli.OutputFormat{cli.OutputFormatText}
}

func (a *CLIArgumentsBookmarksDelete) ShortHelp() [][]string {
	return [][]string{
		{"ffsclient bookmarks delete <id>", "Delete the specified bookmark"},
	}
}

func (a *CLIArgumentsBookmarksDelete) FullHelp() []string {
	return []string{
		"$> ffsclient bookmarks delete <id> [--hard]",
		"",
		"Delete the specific bookmark from the server",
		"(Also modified the parent record and removes the <id> from its children)",
	}
}

func (a *CLIArgumentsBookmarksDelete) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	a.RecordID = positionalArgs[0]

	for _, arg := range optionArgs {
		return fferr.DirectOutput.New("Unknown argument: " + arg.Key)
	}

	return nil
}

func (a *CLIArgumentsBookmarksDelete) Execute(ctx *cli.FFSContext) error {
	ctx.PrintVerbose("[Delete Bookmark]")
	ctx.PrintVerbose("")
	ctx.PrintVerboseKV("RecordID", a.RecordID)

	// ========================================================================

	client, session, err := a.InitClient(ctx)
	if err != nil {
		return err
	}

	// ========================================================================

	ctx.PrintVerboseHeader("[0] Find bookmark")

	record, found, err := a.findBookmarkRecord(ctx, client, session, a.RecordID)
	if err != nil {
		return err
	}

	if !found {
		return fferr.NewDirectOutput(consts.ExitcodePasswordNotFound, "Record not found")
	}

	ctx.PrintVerboseHeader("[1] Get parent")

	parent, err := client.GetRecord(ctx, session, consts.CollectionBookmarks, record.ParentID, true)
	parentFound := true
	if errorx.IsOfType(err, fferr.Request404) {
		parentFound = false
		ctx.PrintVerbose(fmt.Sprintf("No parent found (parent-id := %s)", record.ParentID))
	} else if err != nil {
		return err
	}

	ctx.PrintVerboseHeader("[2] Delete Record " + record.ID)

	var parentRecord models.BookmarkRecord
	if parentFound {
		parentRecord, err = models.UnmarshalBookmark(ctx, parent)
		if err != nil {
			return err
		}
	}

	err = client.SoftDeleteRecord(ctx, session, consts.CollectionBookmarks, record.ID)
	if err != nil {
		return err
	}

	if parentFound {
		ctx.PrintVerboseHeader("[3] Update parent " + parentRecord.ID)

		newChildren := make([]string, 0, len(parentRecord.Children))
		for _, v := range parentRecord.Children {
			if v != record.ID {
				newChildren = append(newChildren, v)
			} else {
				ctx.PrintVerbose("Remove child-entry: " + v)
			}
		}

		plainpayload, err := langext.PatchJson(parent.DecodedData, "children", newChildren)
		if err != nil {
			return fferr.DirectOutput.Wrap(err, "failed to patch payload of parent")
		}

		payload, err := client.EncryptPayload(ctx, session, consts.CollectionBookmarks, string(plainpayload))
		if err != nil {
			return err
		}

		update := models.RecordUpdate{
			ID:      parent.ID,
			Payload: langext.Ptr(payload),
		}

		err = client.PutRecord(ctx, session, consts.CollectionBookmarks, update, false, false)
		if err != nil {
			return err
		}
	}

	// ========================================================================

	if langext.Coalesce(ctx.Opt.Format, cli.OutputFormatText) != cli.OutputFormatText {
		return fferr.NewDirectOutput(consts.ExitcodeUnsupportedOutputFormat, "Unsupported output-format: "+ctx.Opt.Format.String())
	}

	ctx.PrintPrimaryOutput("Bookmark " + a.RecordID + " marked as deleted")
	return nil
}
