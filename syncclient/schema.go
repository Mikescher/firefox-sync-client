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

type registerDeviceRequestSchema struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type signCertRequestSchemaPKey struct {
	Algorithm string `json:"algorithm"`
	P         string `json:"p"`
	Q         string `json:"q"`
	G         string `json:"g"`
	Y         string `json:"y"`
}

type signCertRequestSchema struct {
	PublicKey signCertRequestSchemaPKey `json:"publicKey"`
	Duration  int64                     `json:"duration"`
}

type signCertResponseSchema struct {
	Certificate string `json:"cert"`
}

type hawkCredResponseSchema struct {
	ID            string `json:"id"`
	Key           string `json:"key"`
	UID           int64  `json:"uid"`
	APIEndpoint   string `json:"api_endpoint"`
	Duration      int64  `json:"duration"`
	HashAlgorithm string `json:"hashalg"`
	HashedFxAUID  string `json:"hashed_fxa_uid"`
	NodeType      string `json:"node_type"`
}

type collectionsInfoResponseSchema map[string]float64

type collectionsCountResponseSchema map[string]int

type collectionsUsageResponseSchema map[string]float64

type getRecordSchema struct {
	ID       string  `json:"id"`
	Modified float64 `json:"modified"`
	Payload  string  `json:"payload"`
}

type payloadSchema struct {
	Ciphertext string `json:"ciphertext"`
	IV         string `json:"IV"`
	HMAC       string `json:"hmac"`
}

type cryptoKeysSchema struct {
	Default     []string            `json:"default"`
	Collections map[string][]string `json:"collections"`
	Collection  string              `json:"collection"`
}

type listRecordsIDsResponseSchema []string

type listRecordsResponseSchema []recordsResponseSchema

type recordsResponseSchema struct {
	ID        string  `json:"id"`
	Modified  float64 `json:"modified"`
	Payload   string  `json:"payload"`
	SortIndex int64   `json:"sortIndex"`
}
