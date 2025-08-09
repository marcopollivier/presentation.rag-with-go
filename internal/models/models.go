package models

import "time"

// Document representa um documento no sistema RAG
type Document struct {
	ID       string            `json:"id"`
	Content  string            `json:"content"`
	Metadata map[string]string `json:"metadata"`
	Source   string            `json:"source"`
	Created  time.Time         `json:"created"`
}

// QueryRequest representa uma requisição de busca
type QueryRequest struct {
	Query     string  `json:"query" binding:"required"`
	TopK      int     `json:"top_k,omitempty"`
	Threshold float32 `json:"threshold,omitempty"`
}

// QueryResponse representa a resposta de uma busca RAG
type QueryResponse struct {
	Answer           string             `json:"answer"`
	RelevantDocs     []RelevantDocument `json:"relevant_docs"`
	ProcessingTimeMs int64              `json:"processing_time_ms"`
}

// RelevantDocument representa um documento relevante encontrado
type RelevantDocument struct {
	Document Document `json:"document"`
	Score    float32  `json:"score"`
}

// IndexRequest representa uma requisição para indexar documentos
type IndexRequest struct {
	Documents []Document `json:"documents" binding:"required"`
}

// IndexResponse representa a resposta da indexação
type IndexResponse struct {
	Success        bool     `json:"success"`
	IndexedCount   int      `json:"indexed_count"`
	FailedDocs     []string `json:"failed_docs,omitempty"`
	ProcessingTime string   `json:"processing_time"`
}
