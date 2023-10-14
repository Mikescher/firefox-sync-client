package consts

import (
	"errors"
	"gogs.mikescher.com/BlackForestBytes/goext/langext"
	"runtime/debug"
)

const FFSCLIENT_VERSION = "1.6.0"

type BuildInfo struct {
	VCS         *string
	VCSTime     *string
	VCSRevision *string
	VCSModified *string
}

func ReadBuildInfo() (BuildInfo, error) {

	rbi, ok := debug.ReadBuildInfo()
	if !ok {
		return BuildInfo{}, errors.New("Failed to read BuildInfo")
	}

	return BuildInfo{
		VCS:         getBuildInfoSetting(rbi, "vcs"),
		VCSTime:     getBuildInfoSetting(rbi, "vcs.time"),
		VCSRevision: getBuildInfoSetting(rbi, "vcs.revision"),
		VCSModified: getBuildInfoSetting(rbi, "vcs.modified"),
	}, nil
}

func getBuildInfoSetting(rbi *debug.BuildInfo, key string) *string {
	for _, v := range rbi.Settings {
		if v.Key == key {
			return langext.Ptr(v.Value)
		}
	}
	return nil
}
