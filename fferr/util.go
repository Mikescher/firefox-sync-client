package fferr

import (
	"fmt"
	"github.com/joomcode/errorx"
)

func FormatError(err error, verbose bool) string {
	//errx := errorx.Cast(err)
	//if errx == nil {
	//	return err.Error()
	//}

	sub := err
	for sub != nil {

		errx := errorx.Cast(sub)
		if errx == nil {
			break
		}

		if uw := errx.Unwrap(); uw != nil {
			sub = uw
			continue
		}

		if errx.Type() == DirectOutput {
			if verbose {
				return fmt.Sprintf("%s\n\n%+v", errx.Message(), err)
			} else {
				return errx.Message()
			}
		}

		sub = errx.Cause()
	}

	if verbose {
		return fmt.Sprintf("%+v", err)
	} else {
		return err.Error()
	}
}
