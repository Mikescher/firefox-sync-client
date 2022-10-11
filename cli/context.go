package cli

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"ffsyncclient/fferr"
	"ffsyncclient/langext"
	"ffsyncclient/langext/term"
	"fmt"
	"github.com/joomcode/errorx"
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

func (c FFSContext) PrintPrimaryOutputTable(data [][]string, header bool) {
	if c.Opt.Quiet {
		return
	}

	if len(data) == 0 {
		return
	}

	c.PrintPrimaryOutputTableExt(data, header, langext.IntRange(0, len(data[0])))
}

func (c FFSContext) PrintPrimaryOutputTableExt(data [][]string, header bool, columnFilter []int) {
	if c.Opt.Quiet {
		return
	}

	if len(data) == 0 {
		return
	}

	lens := make([]int, len(data[0]))
	for _, row := range data {
		for i, cell := range row {
			lens[i] = langext.Max(lens[i], len([]rune(cell)))
		}
	}

	for rowidx := range data {

		{
			rowstr := ""
			for ic, colidx := range columnFilter {
				if ic > 0 {
					rowstr += "    "
				}
				rowstr += langext.StrRunePadRight(data[rowidx][colidx], " ", lens[colidx])
			}
			c.printPrimaryRaw(rowstr + "\n")
		}

		if rowidx == 0 && header {
			rowstr := ""
			for ic, colidx := range columnFilter {
				if ic > 0 {
					rowstr += "    "
				}
				rowstr += langext.StrRunePadRight("", "-", lens[colidx])
			}
			c.printPrimaryRaw(rowstr + "\n")
		}

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
		writeStderr(term.Red(msg))
	} else {
		writeStderr(msg)
	}
}

func (c FFSContext) printVerboseRaw(msg string) {
	if c.Opt.Quiet {
		return
	}

	if c.Opt.OutputColor {
		writeStdout(term.Gray(msg))
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
