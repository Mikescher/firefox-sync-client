package impl

import (
	"encoding/base64"
	"ffsyncclient/cli"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"ffsyncclient/langext"
	"ffsyncclient/models"
	"ffsyncclient/syncclient"
	"fmt"
	"github.com/joomcode/errorx"
	"strings"
)

type CLIArgumentsBookmarksBase struct {
}

func NewCLIArgumentsBookmarksBase() *CLIArgumentsBookmarksBase {
	return &CLIArgumentsBookmarksBase{}
}

func (a *CLIArgumentsBookmarksBase) Mode() cli.Mode {
	return cli.ModeBookmarksBase
}

func (a *CLIArgumentsBookmarksBase) PositionArgCount() (*int, *int) {
	return nil, nil
}

func (a *CLIArgumentsBookmarksBase) ShortHelp() [][]string {
	return nil
}

func (a *CLIArgumentsBookmarksBase) FullHelp() []string {
	r := []string{
		"$> ffsclient bookmarks (list|delete|create|update)",
		"======================================================",
		"",
	}
	for _, v := range ListSubcommands(a.Mode()) {
		r = append(r, GetModeImpl(v).FullHelp()...)
		r = append(r, "")
	}

	return r
}

func (a *CLIArgumentsBookmarksBase) Init(positionalArgs []string, optionArgs []cli.ArgumentTuple) error {
	return fferr.DirectOutput.New("ffsclient bookmarks must be called with a subcommand (eg `ffsclient bookmarks list`)")
}

func (a *CLIArgumentsBookmarksBase) Execute(ctx *cli.FFSContext) int {
	return consts.ExitcodeError
}

type CLIArgumentsBookmarksUtil struct{}

func (a *CLIArgumentsBookmarksUtil) filterDeleted(ctx *cli.FFSContext, records []models.BookmarkRecord, includeDeleted bool, onlyDeleted bool, bmtype *[]models.BookmarkType) []models.BookmarkRecord {
	result := make([]models.BookmarkRecord, 0, len(records))

	for _, v := range records {
		if v.Deleted && !includeDeleted {
			ctx.PrintVerbose(fmt.Sprintf("Skip entry %v (is deleted and include-deleted == false)", v.ID))
			continue
		}

		if !v.Deleted && onlyDeleted {
			ctx.PrintVerbose(fmt.Sprintf("Skip entry %v (is not deleted and only-deleted == true)", v.ID))
			continue
		}

		if bmtype != nil && !langext.InArray(v.Type, *bmtype) {
			ctx.PrintVerbose(fmt.Sprintf("Skip entry %v (not in type-filter)", v.ID))
			continue
		}

		result = append(result, v)
	}

	return result
}

func (a *CLIArgumentsBookmarksUtil) calculateTree(ctx *cli.FFSContext, bookmarks []models.BookmarkRecord) ([]*models.BookmarkTreeRecord, []*models.BookmarkRecord, []string) {
	processedOkay := make(map[string]*models.BookmarkTreeRecord)
	parentMap := make(map[string]*models.BookmarkTreeRecord)

	roots := make([]*models.BookmarkTreeRecord, 0)

	// create tree(s) from root nodes
	i := 0
	for ; ; i++ {
		changed := false

		for _, v := range bookmarks {
			if _, ok := processedOkay[v.ID]; ok {
				continue
			}

			if v.ParentID == "" || v.ParentID == "places" {
				record := models.BookmarkTreeRecord{BookmarkRecord: v, ResolvedChildren: make([]*models.BookmarkTreeRecord, 0)}
				roots = append(roots, &record)
				processedOkay[v.ID] = &record
				if v.Type == models.BookmarkTypeFolder || v.Type == models.BookmarkTypeLivemark {
					for _, cid := range v.Children {
						parentMap[cid] = &record
					}
				}
				changed = true
				continue
			}

			if parent, ok := parentMap[v.ID]; ok {
				record := models.BookmarkTreeRecord{BookmarkRecord: v, ResolvedChildren: make([]*models.BookmarkTreeRecord, 0)}
				parent.ResolvedChildren = append(parent.ResolvedChildren, &record)
				processedOkay[v.ID] = &record
				if v.Type == models.BookmarkTypeFolder || v.Type == models.BookmarkTypeLivemark {
					for _, cid := range v.Children {
						parentMap[cid] = &record
					}
				}
				changed = true
				continue
			}
		}

		if !changed {
			break
		}
	}
	ctx.PrintVerbose(fmt.Sprintf("Build boookmark-tree after %d iterations (Processed %d/%d with %d roots)", i, len(processedOkay), len(bookmarks), len(roots)))

	missing := make(map[string]bool, 0)

	// properly sort children
	for _, record := range processedOkay {
		if record.Type == models.BookmarkTypeFolder || record.Type == models.BookmarkTypeLivemark {
			newchildren := make([]*models.BookmarkTreeRecord, 0, len(record.ResolvedChildren))
			for _, childid := range record.Children {
				child, ok := langext.ArrFirst(record.ResolvedChildren, func(p *models.BookmarkTreeRecord) bool { return p.ID == childid })
				if ok {
					newchildren = append(newchildren, child)
				} else {
					ctx.PrintVerbose(fmt.Sprintf("[Warn] the bookmark<%s> record %s references a child '%s' that was not found", record.Type, record.ID, childid))
					missing[childid] = true
				}
			}
			record.ResolvedChildren = newchildren
		}
	}

	for _, record := range roots {
		if record.ParentID != "" {
			ctx.PrintVerbose(fmt.Sprintf("[Warn] the bookmark<%s> record %s references a parent '%s' that was not found", record.Type, record.ID, record.ParentID))
			missing[record.ParentID] = true
		}
	}

	// find missing
	unref := make([]*models.BookmarkRecord, 0)
	for _, v := range bookmarks {
		if _, ok := processedOkay[v.ID]; !ok {
			ctx.PrintVerbose(fmt.Sprintf("[Warn] the bookmark record %s is not a root-node but wasn't found in any parent (self.ParentID/Name := ['%s', '%s'])", v.ID, v.ParentID, v.ParentName))

			unref = append(unref, &v)
			continue
		}
	}

	return roots, unref, langext.MapKeyArr(missing)
}

func (a *CLIArgumentsBookmarksUtil) findBookmarkRecord(ctx *cli.FFSContext, client *syncclient.FxAClient, session syncclient.FFSyncSession, query string) (models.BookmarkRecord, bool, error) {

	record, err := client.GetRecord(ctx, session, consts.CollectionBookmarks, query, true)
	if err != nil && errorx.IsOfType(err, fferr.Request404) {
		return models.BookmarkRecord{}, false, nil
	}
	if err != nil {
		return models.BookmarkRecord{}, false, errorx.Decorate(err, "failed to query record")
	}

	bmrec, err := models.UnmarshalBookmark(ctx, record)
	if err != nil {
		return models.BookmarkRecord{}, false, errorx.Decorate(err, "failed to decode password-record")
	}

	return bmrec, true, nil

}

func (a *CLIArgumentsBookmarksUtil) newBookmarkID() string {
	// BSO ids must only contain printable ASCII characters. They should be exactly 12 base64-urlsafe characters
	v := base64.RawURLEncoding.EncodeToString(langext.RandBytes(32))[0:12]
	if v[0] == '-' {
		// it is annoying when the ID starts with an '-', so it's nice to prevent it as much as possible
		v = "A" + v[1:]
	}
	return v
}

func (a *CLIArgumentsBookmarksUtil) calculateParent(ctx *cli.FFSContext, client *syncclient.FxAClient, session syncclient.FFSyncSession, newid string, parentid string, pos int) (models.BookmarkRecord, string, error, int) {
	ctx.PrintVerbose("Query parent by ID")

	record, err := client.GetRecord(ctx, session, consts.CollectionBookmarks, parentid, true)
	if err != nil && errorx.IsOfType(err, fferr.Request404) {
		return models.BookmarkRecord{}, "", fferr.DirectOutput.Wrap(err, fmt.Sprintf("parent-record with ID '%s' not found", parentid)), consts.ExitcodeRecordNotFound
	}
	if err != nil {
		return models.BookmarkRecord{}, "", errorx.Decorate(err, "failed to query parent-record"), consts.ExitcodeError
	}

	bmrec, err := models.UnmarshalBookmark(ctx, record)
	if err != nil {
		return models.BookmarkRecord{}, "", errorx.Decorate(err, "failed to decode bookmark-record"), consts.ExitcodeError
	}

	if bmrec.Type != models.BookmarkTypeFolder {
		return models.BookmarkRecord{}, "", fferr.DirectOutput.New("The parent record must be a folder"), consts.ExitcodeParentNotAFolder
	}

	children := langext.ForceArray(bmrec.Children)

	normpos := pos

	if normpos < 0 {
		normpos = len(children) + normpos + 1
	}

	ctx.PrintVerboseKV("Position", pos)
	ctx.PrintVerboseKV("Parent<old>.children.len", len(children))
	ctx.PrintVerboseKV("Position-normalized", normpos)

	ctx.PrintVerboseKV("Parent<old>.children", strings.Join(children, ", "))

	if normpos == len(children) {
		children = append(children, newid)
	} else if 0 <= normpos && normpos < len(children) {
		children = append(children[:normpos+1], children[normpos:]...)
		children[normpos] = newid
	} else {
		return models.BookmarkRecord{}, "", fferr.DirectOutput.New(fmt.Sprintf("The parent record [%d..%d] does not have an index %d (%d)", 0, len(children), pos, normpos)), consts.ExitcodeInvalidPosition
	}

	ctx.PrintVerboseKV("Parent<new>.children", strings.Join(children, ", "))

	newPlainPayload, err := langext.PatchJson(record.DecodedData, "children", children)
	if err != nil {
		return models.BookmarkRecord{}, "", errorx.Decorate(err, "failed to patch parent-record data"), consts.ExitcodeError
	}
	bmrec.Children = children

	return bmrec, string(newPlainPayload), nil, 0

}
