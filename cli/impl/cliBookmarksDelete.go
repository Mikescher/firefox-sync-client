package impl

import (
	"encoding/json"
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"ffsyncclient/langext"
	"ffsyncclient/models"
	"ffsyncclient/syncclient"
	"fmt"
	"github.com/joomcode/errorx"
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

func (a *CLIArgumentsBookmarksDelete) Execute(ctx *cli.FFSContext) int {
	ctx.PrintVerbose("[Delete Bookmark]")
	ctx.PrintVerbose("")
	ctx.PrintVerboseKV("RecordID", a.RecordID)

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

	record, found, err := a.findBookmarkRecord(ctx, client, session, a.RecordID)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	if !found {
		ctx.PrintErrorMessage("Record not found")
		return consts.ExitcodePasswordNotFound
	}

	parent, err := client.GetRecord(ctx, session, consts.CollectionBookmarks, record.ParentID, true)
	parentFound := true
	if errorx.IsOfType(err, fferr.Request404) {
		parentFound = false
		ctx.PrintVerbose(fmt.Sprintf("No parent found (parent-id := %s)", record.ParentID))
	} else if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	ctx.PrintVerbose("Delete Record " + record.ID)

	var parentRecord models.BookmarkRecord
	if parentFound {
		parentRecord, err = models.UnmarshalBookmark(ctx, parent)
		if err != nil {
			ctx.PrintFatalError(err)
			return consts.ExitcodeError
		}
	}

	err = client.SoftDeleteRecord(ctx, session, consts.CollectionBookmarks, record.ID)
	if err != nil {
		ctx.PrintFatalError(err)
		return consts.ExitcodeError
	}

	if parentFound {
		ctx.PrintVerbose("Modify Record Parent '" + parent.ID + "'")

		newChildren := make([]string, 0, len(parentRecord.Children))
		for _, v := range parentRecord.Children {
			if v != record.ID {
				newChildren = append(newChildren, v)
			} else {
				ctx.PrintVerbose("Remove child-entry: " + v)
			}
		}

		var jsonpayload langext.H
		err = json.Unmarshal(parent.DecodedData, &jsonpayload)
		if err != nil {
			ctx.PrintFatalError(fferr.DirectOutput.Wrap(err, "failed to unmarshal"))
			return consts.ExitcodeError
		}

		jsonpayload["children"] = newChildren

		plainpayload, err := json.Marshal(jsonpayload)
		if err != nil {
			ctx.PrintFatalError(fferr.DirectOutput.Wrap(err, "failed to re-marshal payload"))
			return consts.ExitcodeError
		}

		payload, err := client.EncryptPayload(ctx, session, consts.CollectionBookmarks, string(plainpayload))
		if err != nil {
			ctx.PrintFatalError(err)
			return consts.ExitcodeError
		}

		update := models.RecordUpdate{
			ID:      parent.ID,
			Payload: langext.Ptr(payload),
		}

		err = client.PutRecord(ctx, session, consts.CollectionBookmarks, update, false, false)
		if err != nil {
			ctx.PrintFatalError(err)
			return consts.ExitcodeError
		}
	}

	// ========================================================================

	if langext.Coalesce(ctx.Opt.Format, cli.OutputFormatText) != cli.OutputFormatText {
		ctx.PrintFatalMessage("Unsupported output-format: " + ctx.Opt.Format.String())
		return consts.ExitcodeUnsupportedOutputFormat
	}

	ctx.PrintPrimaryOutput("Bookmark " + a.RecordID + " deleted")
	return 0
}
