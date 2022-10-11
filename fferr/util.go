package fferr

import (
	"ffsyncclient/consts"
	"fmt"
	"github.com/joomcode/errorx"
)

func GetDirectOutput(err error) *errorx.Error {

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
			return errx
		}

		sub = errx.Cause()
	}

	return nil
}

func FormatError(err error, verbose bool) (string, bool) {
	extraSuffix := ""
	if xerr := errorx.Cast(err); xerr != nil {
		if prop, ok := xerr.Property(ExtraData); ok {
			extraSuffix = fmt.Sprintf("\n[Extra]: %v", prop)
		}
	}

	if errx := GetDirectOutput(err); errx != nil {

		empty := false
		if prop, ok := errx.Property(EmptyMessage); ok {
			if pval, ok := prop.(bool); ok {
				empty = pval
			}
		}

		if verbose {

			return fmt.Sprintf("%s\n\n%+v%s", errx.Message(), err, extraSuffix), empty

		} else {
			return errx.Message(), empty
		}
	}

	if verbose {
		return fmt.Sprintf("%+v%s", err, extraSuffix), false
	} else {
		return err.Error(), false
	}
}

func GetExitCode(err error, fallback consts.FFExitCode) consts.FFExitCode {
	sub := err
	for sub != nil {

		errx := errorx.Cast(sub)
		if errx == nil {
			break
		}

		if ec, ok := errx.Property(Exitcode); ok {
			if eci, ok := ec.(consts.FFExitCode); ok {
				return eci
			}
		}

		if uw := errx.Unwrap(); uw != nil {
			sub = uw
			continue
		}

		sub = errx.Cause()
	}

	return fallback
}

func NewDirectOutput(excode consts.FFExitCode, msg string) error {
	return DirectOutput.New(msg).WithProperty(Exitcode, excode)
}

func WrapDirectOutput(err error, excode consts.FFExitCode, msg string) error {
	return DirectOutput.Wrap(err, msg).WithProperty(Exitcode, excode)
}

func NewEmpty(excode consts.FFExitCode) error {
	return DirectOutput.New("").WithProperty(Exitcode, excode).WithProperty(EmptyMessage, true)
}
