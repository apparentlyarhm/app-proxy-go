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
