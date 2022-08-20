package syncclient

type loginRequestSchema struct {
	Email  string `json:"email"`
	AuthPW string `json:"authPW"`
	Reason string `json:"reason"`
}

type loginResponseSchema struct {
	UserID         string `json:"uid"`
	SessionToken   string `json:"sessionToken"`
	AuthAt         int64  `json:"authAt"`
	MetricsEnabled bool   `json:"metricsEnabled"`
	KeyFetchToken  string `json:"keyFetchToken"`
	Verified       bool   `json:"verified"`
}

type keysResponseSchema struct {
	Bundle string `json:"bundle"`
}
