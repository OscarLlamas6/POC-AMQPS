package models

import "time"

type Message struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Metadata  Metadata  `json:"metadata"`
}

type Metadata struct {
	Source string `json:"source"`
	Type   string `json:"type"`
}

type MessageRequest struct {
	Content  string   `json:"content" binding:"required"`
	Metadata Metadata `json:"metadata"`
}

type MessageResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	ID      string `json:"id,omitempty"`
}
