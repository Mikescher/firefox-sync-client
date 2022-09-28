package fferr

import "github.com/joomcode/errorx"

var (
	FFSyncErrors = errorx.NewNamespace("ffsync")
)

var (
	Request404           = FFSyncErrors.NewType("http_404")
	Request400           = FFSyncErrors.NewType("http_400")
	DirectOutput         = FFSyncErrors.NewType("direct_out")
	UnmarshalConsistency = FFSyncErrors.NewType("unmarshal-consistency")
)
