package syncclient

type loginRequestSchema struct {
	Email  string `json:"email"`
	AuthPW string `json:"authPW"`
	Reason string `json:"reason"`
}

type loginResponseSchema struct {
	UserID             string `json:"uid"`
	SessionToken       string `json:"sessionToken"`
	AuthAt             int64  `json:"authAt"`
	MetricsEnabled     bool   `json:"metricsEnabled"`
	KeyFetchToken      string `json:"keyFetchToken"`
	Verified           bool   `json:"verified"`
	VerificationMethod string `json:"verificationMethod"`
}

type totpVerifyRequestSchema struct {
	Code    string `json:"code"`
	Service string `json:"service"`
}

type totpVerifyResponseSchema struct {
	Success bool `json:"success"`
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

type payloadSchema struct {
	Ciphertext string `json:"ciphertext"`
	IV         string `json:"IV"`
	HMAC       string `json:"hmac"`
}

type deletedPayloadData struct {
	ID      string `json:"id"`
	Deleted bool   `json:"deleted"`
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
	TTL       *int64  `json:"ttl"`
}

type recordsRequestSchema struct {
	ID        *string `json:"id"`
	SortIndex *int64  `json:"sortindex,omitempty"`
	Payload   *string `json:"payload,omitempty"`
	TTL       *int64  `json:"ttl,omitempty"`
}

type sessionStatusResponseSchema struct {
	State  string `json:"state"`
	UserID string `json:"uid"`
}

type oauthTokenRequestSchema struct {
	GrantType    string `json:"grant_type"`
	AccessType   string `json:"access_type,omitempty"`
	ClientID     string `json:"client_id"`
	Scope        string `json:"scope"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

type oauthTokenResponseSchema struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in"`
	AuthAt       int    `json:"auth_at"`
	RefreshToken string `json:"refresh_token"`
}

type scopedKeyDataRequestSchema struct {
	ClientID string `json:"client_id"`
	Scope    string `json:"scope"`
}

type scopedKeyDataResponseSchema map[string]struct {
	Identifier           string `json:"identifier"`
	KeyRotationSecret    string `json:"keyRotationSecret"`
	KeyRotationTimestamp int64  `json:"keyRotationTimestamp"`
}
