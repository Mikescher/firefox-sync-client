package consts

type FFExitCode struct {
	Raw int
}

var (
	ExitcodeOkay = FFExitCode{0}
)

var (
	ExitcodeError                   = FFExitCode{60}
	ExitcodePanic                   = FFExitCode{61}
	ExitcodeNoArguments             = FFExitCode{62}
	ExitcodeCLIParse                = FFExitCode{63}
	ExitcodeNoLogin                 = FFExitCode{64}
	ExitcodeUnsupportedOutputFormat = FFExitCode{65}
	ExitcodeRecordNotFound          = FFExitCode{66}
)

var (
	ExitcodeInvalidSession            = FFExitCode{81}
	ExitcodePasswordNotFound          = FFExitCode{82}
	ExitcodeParentNotAFolder          = FFExitCode{83}
	ExitcodeInvalidPosition           = FFExitCode{84}
	ExitcodeBookmarkFieldNotSupported = FFExitCode{85}
)
