package parser

import (
	"ffsyncclient/cli"
	"ffsyncclient/cli/impl"
	"ffsyncclient/consts"
	"ffsyncclient/fferr"
	"ffsyncclient/langext"
	"fmt"
	"github.com/joomcode/errorx"
	"os"
	"strings"
	"time"
)

func ParseCommandline() (cli.Verb, cli.Options, error) {
	v, o, err := parseCommandlineInternal()
	if err != nil {
		return nil, cli.Options{}, errorx.Decorate(err, "failed to parse commandline")
	}
	return v, o, nil
}

func parseCommandlineInternal() (cli.Verb, cli.Options, error) {
	var err error

	unprocessedArgs := os.Args[1:]

	// Process special cases

	if len(unprocessedArgs) == 0 {
		return &impl.CLIArgumentsHelp{Extra: "ffsclient: missing arguments", ExitCode: consts.ExitcodeNoArguments}, cli.Options{}, nil
	}

	if unprocessedArgs[0] == "-v" {
		return &impl.CLIArgumentsVersion{}, cli.Options{}, nil
	}
	if unprocessedArgs[0] == "--version" {
		return &impl.CLIArgumentsVersion{}, cli.Options{}, nil
	}
	if unprocessedArgs[0] == "-h" {
		return &impl.CLIArgumentsHelp{}, cli.Options{}, nil
	}
	if unprocessedArgs[0] == "--help" {
		return &impl.CLIArgumentsHelp{}, cli.Options{}, nil
	}

	if strings.HasPrefix(unprocessedArgs[0], "-") {
		return nil, cli.Options{}, errorx.InternalError.New("Failed to parse commandline arguments") // no verb
	}

	// Get verb (sub_routine)

	verbArg, rawVerb, verbLen, found := impl.ParseSubcommand(unprocessedArgs)
	if !found {
		return nil, cli.Options{}, fferr.DirectOutput.New(fmt.Sprintf("Unknown Subcommand '%s'", rawVerb))
	}

	unprocessedArgs = unprocessedArgs[verbLen:]

	positionalArguments := make([]string, 0)
	allOptionArguments := make([]cli.ArgumentTuple, 0)

	// Process arguments

	positional := true
	for len(unprocessedArgs) > 0 {
		arg := unprocessedArgs[0]
		unprocessedArgs = unprocessedArgs[1:]

		if !strings.HasPrefix(arg, "-") {
			if !positional {
				return nil, cli.Options{}, fferr.DirectOutput.New("Unknown/Misplaced argument: " + arg)
			}
			positionalArguments = append(positionalArguments, arg)
			continue
		}
		if strings.HasPrefix(arg, "--!arg=") {
			if !positional {
				return nil, cli.Options{}, fferr.DirectOutput.New("Unknown/Misplaced argument: " + arg)
			}
			positionalArguments = append(positionalArguments, arg[7:])
			continue
		}

		positional = false

		if strings.HasPrefix(arg, "--") {

			arg = arg[2:]

			if strings.Contains(arg, "=") {
				key := arg[0:strings.Index(arg, "=")]
				val := arg[strings.Index(arg, "=")+1:]

				if len(key) <= 1 {
					return nil, cli.Options{}, fferr.DirectOutput.New("Unknown/Misplaced argument: " + arg)
				}

				allOptionArguments = append(allOptionArguments, cli.ArgumentTuple{Key: key, Value: langext.Ptr(val)})
				continue
			} else {

				key := arg

				if len(key) <= 1 {
					return nil, cli.Options{}, fferr.DirectOutput.New("Unknown/Misplaced argument: " + arg)
				}

				if len(unprocessedArgs) == 0 || strings.HasPrefix(unprocessedArgs[0], "-") {
					allOptionArguments = append(allOptionArguments, cli.ArgumentTuple{Key: key, Value: nil})
					continue
				} else {
					val := unprocessedArgs[0]
					unprocessedArgs = unprocessedArgs[1:]
					allOptionArguments = append(allOptionArguments, cli.ArgumentTuple{Key: key, Value: langext.Ptr(val)})
					continue
				}

			}

		} else if strings.HasPrefix(arg, "-") {

			arg = arg[1:]

			if len(arg) > 1 {
				for i := 1; i < len(arg); i++ {
					allOptionArguments = append(allOptionArguments, cli.ArgumentTuple{Key: arg[i : i+1], Value: nil})
				}
				continue
			}

			key := arg

			if key == "" {
				return nil, cli.Options{}, fferr.DirectOutput.New("Unknown/Misplaced argument: " + arg)
			}

			if len(unprocessedArgs) == 0 || strings.HasPrefix(unprocessedArgs[0], "-") {
				allOptionArguments = append(allOptionArguments, cli.ArgumentTuple{Key: key, Value: nil})
				continue
			} else {
				val := unprocessedArgs[0]
				unprocessedArgs = unprocessedArgs[1:]
				allOptionArguments = append(allOptionArguments, cli.ArgumentTuple{Key: key, Value: langext.Ptr(val)})
				continue
			}

		} else {
			return nil, cli.Options{}, fferr.DirectOutput.New("Unknown/Misplaced argument: " + arg)
		}
	}

	// Process common options

	opt := cli.DefaultCLIOptions()

	optionArguments := make([]cli.ArgumentTuple, 0)

	for _, arg := range allOptionArguments {

		if (arg.Key == "h" || arg.Key == "help") && arg.Value == nil {
			return &impl.CLIArgumentsHelp{Verb: langext.Ptr(verbArg.Mode())}, cli.Options{}, nil
		}

		if arg.Key == "version" && arg.Value == nil {
			return &impl.CLIArgumentsVersion{}, cli.Options{}, nil
		}

		if (arg.Key == "v" || arg.Key == "verbose") && arg.Value == nil {
			opt.Verbose = true
			continue
		}

		if (arg.Key == "q" || arg.Key == "quiet") && arg.Value == nil {
			opt.Quiet = true
			continue
		}

		if (arg.Key == "f" || arg.Key == "format") && arg.Value != nil {
			ofmt, found := cli.GetOutputFormat(*arg.Value)
			if !found {
				return nil, cli.Options{}, fferr.DirectOutput.New("Unknown output-format: " + *arg.Value)
			}
			opt.Format = langext.Ptr(ofmt)
			continue
		}

		if (arg.Key == "sessionfile" || arg.Key == "session-file") && arg.Value != nil {
			opt.SessionFilePath = *arg.Value
			continue
		}

		if arg.Key == "auth-server" && arg.Value != nil {
			opt.AuthServerURL = *arg.Value
			continue
		}

		if arg.Key == "token-server" && arg.Value != nil {
			opt.TokenServerURL = *arg.Value
			continue
		}

		if arg.Key == "timezone" && arg.Value != nil {
			loc, err := time.LoadLocation(*arg.Value)
			if err != nil {
				return nil, cli.Options{}, fferr.DirectOutput.New("Unknown timezone: " + *arg.Value)
			}
			opt.TimeZone = loc
			continue
		}

		if arg.Key == "timeformat" && arg.Value != nil {
			opt.TimeFormat = *arg.Value
			continue
		}

		if arg.Key == "color" && arg.Value == nil {
			opt.OutputColor = true
			continue
		}

		if arg.Key == "color" && arg.Value != nil && (strings.ToLower(*arg.Value) == "true" || *arg.Value == "1") {
			opt.OutputColor = true
			continue
		}

		if arg.Key == "color" && arg.Value != nil && (strings.ToLower(*arg.Value) == "false" || *arg.Value == "0") {
			opt.OutputColor = false
			continue
		}

		if arg.Key == "no-color" && arg.Value == nil {
			opt.OutputColor = false
			continue
		}

		if (arg.Key == "o" || arg.Key == "output") && arg.Value != nil {
			opt.OutputFile = langext.Ptr(*arg.Value)
			continue
		}

		if arg.Key == "no-autosave-session" && arg.Value == nil {
			opt.SaveRefreshedSession = false
			continue
		}

		if arg.Key == "force-refresh-session" && arg.Value == nil {
			opt.ForceRefreshSession = false
			continue
		}

		if arg.Key == "no-xml-declaration" && arg.Value == nil {
			opt.NoXMLDeclaration = true
			continue
		}

		if arg.Key == "minimized-json" && arg.Value == nil {
			opt.LinearizeJson = true
			continue
		}

		optionArguments = append(optionArguments, arg)
	}

	posArgLenMin, posArgLenMax := verbArg.PositionArgCount()
	if posArgLenMin != nil && posArgLenMax != nil && *posArgLenMin == *posArgLenMax {
		if len(positionalArguments) < *posArgLenMin {
			return nil, cli.Options{}, fferr.DirectOutput.New(fmt.Sprintf("Not enough arguments for `ffsclient %s` (must be exactly %d)", verbArg.Mode(), *posArgLenMin))
		}
		if len(positionalArguments) > *posArgLenMax {
			if *posArgLenMax == 0 {
				return nil, cli.Options{}, fferr.DirectOutput.New(fmt.Sprintf("Command `ffsclient %s` does not have any subcommands", verbArg.Mode()))
			} else {
				return nil, cli.Options{}, fferr.DirectOutput.New(fmt.Sprintf("Too many arguments for `ffsclient %s` (must be exactly %d)", verbArg.Mode(), *posArgLenMax))
			}
		}
	}
	if posArgLenMin != nil && len(positionalArguments) < *posArgLenMin {
		return nil, cli.Options{}, fferr.DirectOutput.New(fmt.Sprintf("Not enough arguments for `ffsclient %s` (must be at least %d)", verbArg.Mode(), *posArgLenMin))
	}
	if posArgLenMax != nil && len(positionalArguments) > *posArgLenMax {
		return nil, cli.Options{}, fferr.DirectOutput.New(fmt.Sprintf("Too many arguments for `ffsclient %s` (must be at most %d)", verbArg.Mode(), *posArgLenMax))
	}

	err = verbArg.Init(positionalArguments, optionArguments)
	if err != nil {
		return nil, cli.Options{}, errorx.Decorate(err, "failed to init "+verbArg.Mode().String())
	}

	possibleFormats := verbArg.AvailableOutputFormats()
	if opt.Format != nil && !langext.InArray(*opt.Format, possibleFormats) {
		return nil, cli.Options{}, fferr.DirectOutput.New(fmt.Sprintf("The output format '%s' is not supported in this subcommand.\nSupported formats are: %s", *opt.Format, joinOutputFormats(possibleFormats))).WithProperty(fferr.Exitcode, consts.ExitcodeUnsupportedOutputFormat)
	}

	return verbArg, opt, nil
}

func joinOutputFormats(f []cli.OutputFormat) string {
	a := make([]string, 0, len(f))
	for _, v := range f {
		a = append(a, string(v))
	}
	return strings.Join(a, ", ")
}
