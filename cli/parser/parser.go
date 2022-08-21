package parser

import (
	"ffsyncclient/cli"
	"ffsyncclient/cli/impl"
	"ffsyncclient/consts"
	"ffsyncclient/langext"
	"github.com/joomcode/errorx"
	"os"
	"strings"
	"time"
)

func ParseCommandline() (cli.Verb, cli.Options) {
	v, o, err := parseCommandlineInternal()
	if err != nil {
		return &impl.CLIArgumentsHelp{Extra: err.Error(), ExitCode: consts.ExitcodeCLIParse}, cli.Options{}
	}
	return v, o
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

	verb := unprocessedArgs[0]
	unprocessedArgs = unprocessedArgs[1:]

	verbArg, found := getVerb(verb)
	if !found {
		return nil, cli.Options{}, errorx.InternalError.New("Unknown command: " + verb)
	}

	positionalArguments := make([]string, 0)
	allOptionArguments := make([]cli.ArgumentTuple, 0)

	// Process arguments

	positional := true
	for len(unprocessedArgs) > 0 {
		arg := unprocessedArgs[0]
		unprocessedArgs = unprocessedArgs[1:]

		if !strings.HasPrefix(arg, "-") {
			if !positional {
				return nil, cli.Options{}, errorx.InternalError.New("Unknown/Misplaced argument: " + arg)
			}
			positionalArguments = append(positionalArguments, arg)
			continue
		}

		positional = false

		if strings.HasPrefix(arg, "--") {

			arg = arg[2:]

			if strings.Contains(arg, "=") {
				key := arg[0:strings.Index(arg, "=")]
				val := arg[strings.Index(arg, "=")+1:]

				if len(key) <= 1 {
					return nil, cli.Options{}, errorx.InternalError.New("Unknown/Misplaced argument: " + arg)
				}

				allOptionArguments = append(allOptionArguments, cli.ArgumentTuple{Key: key, Value: langext.Ptr(val)})
				continue
			} else {

				key := arg

				if len(key) <= 1 {
					return nil, cli.Options{}, errorx.InternalError.New("Unknown/Misplaced argument: " + arg)
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
				return nil, cli.Options{}, errorx.InternalError.New("Unknown/Misplaced argument: " + arg)
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
			return nil, cli.Options{}, errorx.InternalError.New("Unknown/Misplaced argument: " + arg)
		}
	}

	// Process common options

	opt := cli.DefaultCLIOptions()

	optionArguments := make([]cli.ArgumentTuple, 0)

	for _, arg := range allOptionArguments {

		if arg.Key == "help" && arg.Value == nil {
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
			opt.Verbose = true
			continue
		}

		if (arg.Key == "f" || arg.Key == "format") && arg.Value != nil {
			fmt, found := cli.GetOutputFormat(*arg.Value)
			if !found {
				return nil, cli.Options{}, errorx.InternalError.New("Unknown format: " + *arg.Value)
			}
			opt.Format = langext.Ptr(fmt)
			continue
		}

		if (arg.Key == "c" || arg.Key == "conf" || arg.Key == "config") && arg.Value != nil {
			opt.ConfigFilePath = *arg.Value
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
				return nil, cli.Options{}, errorx.InternalError.New("Unknown timezone: " + *arg.Value)
			}
			opt.TimeZone = loc
			continue
		}

		if arg.Key == "timeformat" && arg.Value != nil {
			opt.TimeFormat = *arg.Value
			continue
		}

		if arg.Key == "color" && arg.Value == nil {
			opt.OutputColor = langext.Ptr(true)
			continue
		}

		if arg.Key == "color" && arg.Value != nil && (strings.ToLower(*arg.Value) == "true" || *arg.Value == "1") {
			opt.OutputColor = langext.Ptr(true)
			continue
		}

		if arg.Key == "color" && arg.Value != nil && (strings.ToLower(*arg.Value) == "false" || *arg.Value == "0") {
			opt.OutputColor = langext.Ptr(false)
			continue
		}

		if arg.Key == "no-color" && arg.Value == nil {
			opt.OutputColor = langext.Ptr(false)
			continue
		}

		optionArguments = append(optionArguments, arg)
	}

	err = verbArg.Init(positionalArguments, optionArguments)
	if err != nil {
		return nil, cli.Options{}, errorx.Decorate(err, "failed to init "+verbArg.Mode().String())
	}

	return verbArg, opt, nil
}
