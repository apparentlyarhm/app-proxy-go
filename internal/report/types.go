package report

type SystemInfo struct {
	Timestamp int64    `json:"timestamp"`
	Arch      string   `json:"arch"`
	OSName    string   `json:"os_name"`
	OSVersion string   `json:"os_version"`
	CPU       string   `json:"cpu"`
	MemoryGB  float64  `json:"memory_gb"`
	GPUs      []string `json:"gpus"`
}

// This struct represents the payload for recording a blog view.
// slug is the blog id, issuedAt is the timestamp of the view, and
// nonce is a random string to prevent replay attacks.

// ViewPayload is the actual internal data inside the Base64 payload
type ViewPayload struct {
	Slug     string `json:"slug"`
	IssuedAt int64  `json:"iat"`
	Nonce    string `json:"nonce"`
}

// the actual HTTP data coming from the blog server.
type ViewRequest struct {
	Payload   string `json:"payload"`
	Signature string `json:"signature"`
}
