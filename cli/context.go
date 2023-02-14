package cli

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"ffsyncclient/fferr"
	"fmt"
	"github.com/joomcode/errorx"
	"gogs.mikescher.com/BlackForestBytes/goext/langext"
	"gogs.mikescher.com/BlackForestBytes/goext/mathext"
	"gogs.mikescher.com/BlackForestBytes/goext/termext"
	"golang.org/x/term"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

type FFSContext struct {
	context.Context
	Opt        Options
	FileHandle *os.File
}

func (c FFSContext) PrintPrimaryOutput(msg string) {
	if c.Opt.Quiet {
		return
	}

	c.printPrimaryRaw(msg + "\n")
}

func (c FFSContext) PrintPrimaryOutputJSON(data any) {
	if c.Opt.Quiet {
		return
	}

	if c.Opt.LinearizeJson {
		msg, err := json.Marshal(data)
		if err != nil {
			panic("failed to marshal output: " + err.Error())
		}

		c.printPrimaryRaw(string(msg) + "\n")
	} else {
		msg, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			panic("failed to marshal output: " + err.Error())
		}

		c.printPrimaryRaw(string(msg) + "\n")
	}

}

func (c FFSContext) PrintPrimaryOutputXML(data any) {
	if c.Opt.Quiet {
		return
	}

	msg, err := xml.MarshalIndent(data, "", "  ")
	if err != nil {
		panic("failed to marshal output: " + err.Error())
	}

	if !c.Opt.NoXMLDeclaration {
		c.printPrimaryRaw("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	}
	c.printPrimaryRaw(string(msg) + "\n")
}

func (c FFSContext) PrintPrimaryOutputTable(data [][]string) {
	if c.Opt.Quiet {
		return
	}

	if len(data) == 0 {
		return
	}

	c.PrintPrimaryOutputTableExt(data, langext.Range(0, len(data[0])))
}

func (c FFSContext) PrintPrimaryOutputTableExt(data [][]string, columnFilter []int) {

	realColFilter := make([]int, 0, len(columnFilter))

	if c.Opt.TableFormatFilter == nil {

		realColFilter = columnFilter

	} else if len(data) == 0 {

		realColFilter = columnFilter

	} else {

		cleanHeader := func(v string) string {
			v = strings.TrimSpace(v)
			v = strings.ToLower(v)
			v = strings.ReplaceAll(v, " ", "")
			return v
		}

		cheaders := langext.ArrMap(data[0], cleanHeader)

		for _, uf := range strings.Split(*c.Opt.TableFormatFilter, ",") {

			cuf := cleanHeader(uf)

			idx := langext.ArrFirstIndex(cheaders, cuf)
			if idx >= 0 && langext.InArray(idx, columnFilter) {

				realColFilter = append(realColFilter, idx)

			}
		}
	}

	c.printPrimaryTable(data, realColFilter, c.Opt.TableFormatTruncate)
}

func (c FFSContext) printPrimaryTable(data [][]string, columnFilter []int, truncate bool) {
	if c.Opt.Quiet {
		return
	}

	if len(data) == 0 {
		return
	}

	colsep := "    "

	lens := make([]int, len(data[0]))
	for _, row := range data {
		for i, cell := range row {
			lens[i] = mathext.Max(lens[i], len([]rune(cell)))
		}
	}

	if truncate && len(data) > 0 {
		filteredLens := make([]int, 0, len(lens))
		filteredHdr := make([]string, 0, len(data[0]))
		for _, colidx := range columnFilter {
			filteredLens = append(filteredLens, lens[colidx])
			filteredHdr = append(filteredHdr, data[0][colidx])
		}

		tableWidth := langext.ArrSum(filteredLens) + (len(filteredLens)-1)*len(colsep)
		termWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
		if err == nil && tableWidth > termWidth {
			filteredLens = c.truncateTableColumns(filteredLens, termWidth, filteredHdr, len(colsep))
			for ic, colidx := range columnFilter {
				lens[colidx] = filteredLens[ic]
			}
		}
	}

	for rowidx := range data {

		{
			rowstr := ""
			for ic, colidx := range columnFilter {
				if ic > 0 {
					rowstr += colsep
				}
				str := langext.StrRunePadRight(data[rowidx][colidx], " ", lens[colidx])
				if truncate {
					str = c.truncTableData(str, lens[colidx])
				}
				rowstr += str
			}
			c.printPrimaryRaw(rowstr + "\n")
		}

		if rowidx == 0 {
			rowstr := ""
			for ic, colidx := range columnFilter {
				if ic > 0 {
					rowstr += colsep
				}
				str := langext.StrRunePadRight("", "-", lens[colidx])
				if truncate {
					str = c.truncTableData(str, lens[colidx])
				}
				rowstr += str
			}
			c.printPrimaryRaw(rowstr + "\n")
		}

	}

}

func (c FFSContext) truncateTableColumns(calcLens []int, termWidth int, header []string, colsepWidth int) []int {

	colCount := len(calcLens)

	remainingWidth := termWidth - colsepWidth*(colCount-1)

	runelen := func(v string) int {
		return len([]rune(v))
	}

	realWidths := make([]int, colCount) // all zero

	// reserve at least header
	for i := 0; i < colCount; i++ {

		realWidths[i] = runelen(header[i])
		remainingWidth -= realWidths[i]

	}

	for remainingWidth > 0 {

		changed := false
		for i := 0; i < colCount; i++ {
			if realWidths[i] >= calcLens[i] {
				continue // already happy
			}
			realWidths[i]++
			remainingWidth--
			changed = true

			if remainingWidth <= 0 {
				return realWidths // early return
			}

		}
		if !changed {
			return realWidths // everyones happy
		}

	}

	return realWidths
}

func (c FFSContext) truncTableData(str string, collen int) string {
	rlen := len([]rune(str))
	if rlen > collen {
		if collen <= 3 {
			return strings.Repeat(".", collen)
		}
		return string(([]rune(str))[0:collen-3]) + "..."
	} else {
		return str
	}
}

func (c FFSContext) PrintPrimaryOutputCSV(data [][]string, tsv bool) {
	if c.Opt.Quiet {
		return
	}

	for _, v := range data {
		buffer := bytes.NewBuffer([]byte{})
		writer := csv.NewWriter(buffer)
		if tsv {
			writer.Comma = '\t'
		}
		err := writer.Write(v)
		writer.Flush()
		if err != nil {
			panic("failed to marshal csv: " + err.Error())
		}
		c.printPrimaryRaw(buffer.String())
	}
}

func (c FFSContext) PrintFatalMessage(msg string) {
	if c.Opt.Quiet {
		return
	}

	c.printErrorRaw(msg + "\n")
}

func (c FFSContext) PrintFatalError(e error) {
	if c.Opt.Quiet {
		return
	}

	errfmt, empty := fferr.FormatError(e, c.Opt.Verbose)
	if empty {
		return
	}

	c.printErrorRaw(errfmt + "\n")
}

func (c FFSContext) PrintErrorMessage(msg string) {
	if c.Opt.Quiet {
		return
	}

	c.printErrorRaw(msg + "\n")
}

func (c FFSContext) PrintVerbose(msg string) {
	if c.Opt.Quiet || !c.Opt.Verbose {
		return
	}

	c.printVerboseRaw(msg + "\n")
}

func (c FFSContext) PrintVerboseHeader(msg string) {
	if c.Opt.Quiet || !c.Opt.Verbose {
		return
	}

	c.printVerboseRaw("\n")
	c.printVerboseRaw("========================================" + "\n")
	c.printVerboseRaw(msg + "\n")
	c.printVerboseRaw("========================================" + "\n")
	c.printVerboseRaw("\n")
}

func (c FFSContext) PrintVerboseKV(key string, vval any) {
	if c.Opt.Quiet || !c.Opt.Verbose {
		return
	}

	termlen := 236
	keylen := 28

	var val = ""
	switch v := vval.(type) {
	case []byte:
		val = hex.EncodeToString(v)
	case string:
		val = v
	case time.Time:
		val = v.In(c.Opt.TimeZone).Format(time.RFC3339Nano)
	default:
		val = fmt.Sprintf("%v", v)
	}

	if len(val) > (termlen-keylen-4) || strings.Contains(val, "\n") {

		c.printVerboseRaw(key + " :=\n" + val + "\n")

	} else {

		padkey := langext.StrPadRight(key, " ", keylen)
		c.printVerboseRaw(padkey + " := " + val + "\n")

	}
}

func (c FFSContext) printPrimaryRaw(msg string) {
	if c.Opt.Quiet {
		return
	}

	if c.FileHandle == nil {
		writeStdout(msg)
	} else {
		_, err := c.FileHandle.WriteString(msg)
		if err != nil {
			panic("failed to write to file: " + err.Error())
		}
	}

}

func (c FFSContext) printErrorRaw(msg string) {
	if c.Opt.Quiet {
		return
	}

	if c.Opt.OutputColor {
		writeStderr(termext.Red(msg))
	} else {
		writeStderr(msg)
	}
}

func (c FFSContext) printVerboseRaw(msg string) {
	if c.Opt.Quiet {
		return
	}

	if c.Opt.OutputColor {
		writeStdout(termext.Gray(msg))
	} else {
		writeStdout(msg)
	}
}

func (c FFSContext) AbsSessionFilePath() (string, error) {
	fp := c.Opt.SessionFilePath

	if fp == "~" {
		usr, err := user.Current()
		if err != nil {
			return "", errorx.Decorate(err, "failed to get current user")
		}
		fp = usr.HomeDir
	} else if strings.HasPrefix(fp, "~/") {
		usr, err := user.Current()
		if err != nil {
			return "", errorx.Decorate(err, "failed to get current user")
		}
		fp = filepath.Join(usr.HomeDir, fp[2:])
	}

	fp, err := filepath.Abs(fp)
	if err != nil {
		return "", errorx.Decorate(err, "failed to parse session filepath")
	}

	return fp, nil
}

func writeStdout(msg string) {
	_, err := os.Stdout.WriteString(msg)
	if err != nil {
		panic("failed to write to stdout: " + err.Error())
	}
}

func writeStderr(msg string) {
	_, err := os.Stderr.WriteString(msg)
	if err != nil {
		panic("failed to write to stdout: " + err.Error())
	}
}

func NewContext(opt Options) (*FFSContext, error) {
	var fileHandle *os.File

	if opt.OutputFile != nil {
		fh, err := os.OpenFile(*opt.OutputFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			return nil, err
		}
		fileHandle = fh
	}

	return &FFSContext{
		Context:    context.Background(),
		Opt:        opt,
		FileHandle: fileHandle,
	}, nil
}

func NewEarlyContext() *FFSContext {
	return &FFSContext{
		Context:    context.Background(),
		Opt:        DefaultCLIOptions(),
		FileHandle: nil,
	}
}

func (c FFSContext) Finish() {
	if c.FileHandle != nil {
		err := c.FileHandle.Close()
		if err != nil {
			c.PrintFatalError(err)
		}
	}
}

func (c FFSContext) ReadStdIn() (string, error) {
	bytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", errorx.Decorate(err, "failed to read from stdin")
	}
	return string(bytes), nil
}
