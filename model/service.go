package model

type Service struct {
	Name        string `json:"name"`
	Image       string `json:"image"`
	Replicas    int    `json:"replicas,omitempty"`
	MemoryLimit int    `json:"memoryLimit,omitempty"`
}
