package qdrant

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/marcopollivier/rag-go-ex01/internal/models"
	"github.com/sirupsen/logrus"
)

type Client struct {
	baseURL        string
	httpClient     *http.Client
	collectionName string
	logger         *logrus.Logger
}

// NewClient cria um novo cliente Qdrant usando HTTP API
func NewClient(host string, port int, collectionName string, logger *logrus.Logger) (*Client, error) {
	baseURL := fmt.Sprintf("http://%s:%d", host, port)

	c := &Client{
		baseURL:        baseURL,
		httpClient:     &http.Client{Timeout: 30 * time.Second},
		collectionName: collectionName,
		logger:         logger,
	}

	// Verificar se a coleção existe, se não, criar
	if err := c.ensureCollection(context.Background()); err != nil {
		return nil, fmt.Errorf("erro ao garantir coleção: %w", err)
	}

	return c, nil
}

// Collection structures
type CollectionConfig struct {
	Vectors struct {
		Size     int    `json:"size"`
		Distance string `json:"distance"`
	} `json:"vectors"`
}

type CreateCollectionRequest struct {
	Vectors struct {
		Size     int    `json:"size"`
		Distance string `json:"distance"`
	} `json:"vectors"`
}

type PointStruct struct {
	ID      string                 `json:"id"`
	Vector  []float32              `json:"vector"`
	Payload map[string]interface{} `json:"payload"`
}

type UpsertRequest struct {
	Points []PointStruct `json:"points"`
}

type SearchRequest struct {
	Vector      []float32 `json:"vector"`
	Limit       int       `json:"limit"`
	Threshold   float32   `json:"score_threshold,omitempty"`
	WithPayload bool      `json:"with_payload"`
}

type SearchResponse struct {
	Result []struct {
		ID      string                 `json:"id"`
		Score   float32                `json:"score"`
		Payload map[string]interface{} `json:"payload"`
	} `json:"result"`
}

type CollectionInfoResponse struct {
	Status string `json:"status"`
	Result struct {
		Status      string `json:"status"`
		PointsCount int    `json:"points_count"`
		Config      struct {
			Params struct {
				Vectors struct {
					Size     int    `json:"size"`
					Distance string `json:"distance"`
				} `json:"vectors"`
			} `json:"params"`
		} `json:"config"`
	} `json:"result"`
}

// ensureCollection garante que a coleção existe
func (c *Client) ensureCollection(ctx context.Context) error {
	c.logger.Infof("Verificando se coleção '%s' existe", c.collectionName)

	// Verificar se coleção existe
	url := fmt.Sprintf("%s/collections/%s", c.baseURL, c.collectionName)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("erro ao verificar coleção: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		c.logger.Infof("Coleção '%s' já existe", c.collectionName)
		return nil
	}

	c.logger.Infof("Criando coleção '%s'", c.collectionName)

	// Criar coleção
	createReq := CreateCollectionRequest{}
	createReq.Vectors.Size = 1536 // OpenAI embeddings
	createReq.Vectors.Distance = "Cosine"

	jsonData, err := json.Marshal(createReq)
	if err != nil {
		return fmt.Errorf("erro ao serializar requisição: %w", err)
	}

	createURL := fmt.Sprintf("%s/collections/%s", c.baseURL, c.collectionName)
	req, err := http.NewRequest("PUT", createURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("erro ao criar requisição: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err = c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("erro ao criar coleção: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("erro ao criar coleção: status %d, body: %s", resp.StatusCode, string(body))
	}

	c.logger.Infof("Coleção '%s' criada com sucesso", c.collectionName)
	return nil
}

// IndexDocument indexa um documento no Qdrant
func (c *Client) IndexDocument(ctx context.Context, doc models.Document, embedding []float32) error {
	c.logger.Debugf("Indexando documento ID: %s", doc.ID)

	// Criar payload
	payload := map[string]interface{}{
		"content": doc.Content,
		"source":  doc.Source,
		"created": doc.Created.Format("2006-01-02T15:04:05Z"),
	}

	// Adicionar metadata
	for key, value := range doc.Metadata {
		payload[fmt.Sprintf("metadata_%s", key)] = value
	}

	point := PointStruct{
		ID:      doc.ID,
		Vector:  embedding,
		Payload: payload,
	}

	upsertReq := UpsertRequest{
		Points: []PointStruct{point},
	}

	jsonData, err := json.Marshal(upsertReq)
	if err != nil {
		return fmt.Errorf("erro ao serializar ponto: %w", err)
	}

	url := fmt.Sprintf("%s/collections/%s/points", c.baseURL, c.collectionName)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("erro ao criar requisição: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.WithError(err).Errorf("Erro ao indexar documento %s", doc.ID)
		return fmt.Errorf("erro ao indexar documento: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("erro ao indexar documento: status %d, body: %s", resp.StatusCode, string(body))
	}

	c.logger.Debugf("Documento %s indexado com sucesso", doc.ID)
	return nil
}

// SearchSimilar busca documentos similares
func (c *Client) SearchSimilar(ctx context.Context, queryEmbedding []float32, topK int, threshold float32) ([]models.RelevantDocument, error) {
	c.logger.Debugf("Buscando %d documentos similares com threshold %.2f", topK, threshold)

	searchReq := SearchRequest{
		Vector:      queryEmbedding,
		Limit:       topK,
		Threshold:   threshold,
		WithPayload: true,
	}

	jsonData, err := json.Marshal(searchReq)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar busca: %w", err)
	}

	url := fmt.Sprintf("%s/collections/%s/points/search", c.baseURL, c.collectionName)
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		c.logger.WithError(err).Error("Erro ao buscar documentos similares")
		return nil, fmt.Errorf("erro ao buscar documentos: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("erro na busca: status %d, body: %s", resp.StatusCode, string(body))
	}

	var searchResponse SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %w", err)
	}

	c.logger.Infof("Resposta do Qdrant: %d resultados encontrados", len(searchResponse.Result))

	var relevantDocs []models.RelevantDocument
	for _, hit := range searchResponse.Result {
		c.logger.Infof("Processando hit ID: %s, Score: %.4f", hit.ID, hit.Score)

		// Extrair dados do payload
		content := ""
		source := ""
		created := time.Time{}
		metadata := make(map[string]string)

		if contentVal, ok := hit.Payload["content"]; ok {
			if str, ok := contentVal.(string); ok {
				content = str
				c.logger.Infof("Content encontrado: %d caracteres", len(content))
			} else {
				c.logger.Warn("Content não é uma string")
			}
		} else {
			c.logger.Warn("Content não encontrado no payload")
		}

		if sourceVal, ok := hit.Payload["source"]; ok {
			if str, ok := sourceVal.(string); ok {
				source = str
			}
		}
		if createdVal, ok := hit.Payload["created"]; ok {
			if str, ok := createdVal.(string); ok {
				if t, err := time.Parse("2006-01-02T15:04:05Z", str); err == nil {
					created = t
				}
			}
		}

		// Extrair metadata
		for key, value := range hit.Payload {
			if len(key) > 9 && key[:9] == "metadata_" {
				metaKey := key[9:]
				if str, ok := value.(string); ok {
					metadata[metaKey] = str
				}
			}
		}

		doc := models.Document{
			ID:       hit.ID,
			Content:  content,
			Source:   source,
			Metadata: metadata,
			Created:  created,
		}

		c.logger.Infof("Documento criado - ID: %s, Content: %d chars, Source: %s", doc.ID, len(doc.Content), doc.Source)

		relevantDocs = append(relevantDocs, models.RelevantDocument{
			Document: doc,
			Score:    hit.Score,
		})
	}

	c.logger.Debugf("Encontrados %d documentos relevantes", len(relevantDocs))
	return relevantDocs, nil
}

// DeleteDocument remove um documento do índice
func (c *Client) DeleteDocument(ctx context.Context, docID string) error {
	c.logger.Debugf("Removendo documento ID: %s", docID)

	deleteReq := map[string]interface{}{
		"points": []string{docID},
	}

	jsonData, err := json.Marshal(deleteReq)
	if err != nil {
		return fmt.Errorf("erro ao serializar delete: %w", err)
	}

	url := fmt.Sprintf("%s/collections/%s/points/delete", c.baseURL, c.collectionName)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("erro ao criar requisição: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.WithError(err).Errorf("Erro ao remover documento %s", docID)
		return fmt.Errorf("erro ao remover documento: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("erro ao remover documento: status %d, body: %s", resp.StatusCode, string(body))
	}

	c.logger.Debugf("Documento %s removido com sucesso", docID)
	return nil
}

// ScrollRequest representa uma requisição de scroll
type ScrollRequest struct {
	Limit       int                    `json:"limit"`
	WithPayload bool                   `json:"with_payload"`
	WithVector  bool                   `json:"with_vector"`
	Offset      *string                `json:"offset,omitempty"`
	Filter      map[string]interface{} `json:"filter,omitempty"`
}

// ScrollResponse representa a resposta de um scroll
type ScrollResponse struct {
	Result struct {
		Points         []PointStruct `json:"points"`
		NextPageOffset *string       `json:"next_page_offset"`
	} `json:"result"`
}

// GetAllDocuments retorna todos os documentos da coleção usando scroll
func (c *Client) GetAllDocuments(ctx context.Context, limit int) ([]models.Document, error) {
	c.logger.Infof("Buscando todos os documentos da coleção '%s'", c.collectionName)

	scrollReq := ScrollRequest{
		Limit:       limit,
		WithPayload: true,
		WithVector:  false,
	}

	jsonData, err := json.Marshal(scrollReq)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar scroll: %w", err)
	}

	url := fmt.Sprintf("%s/collections/%s/points/scroll", c.baseURL, c.collectionName)
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("erro ao fazer scroll: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("erro no scroll: status %d, body: %s", resp.StatusCode, string(body))
	}

	var scrollResponse ScrollResponse
	if err := json.NewDecoder(resp.Body).Decode(&scrollResponse); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %w", err)
	}

	var documents []models.Document
	for _, point := range scrollResponse.Result.Points {
		doc := c.pointToDocument(point)
		documents = append(documents, doc)
	}

	c.logger.Infof("Encontrados %d documentos na coleção", len(documents))
	return documents, nil
}

// GetDocumentsBySource busca documentos por fonte específica
func (c *Client) GetDocumentsBySource(ctx context.Context, source string, limit int) ([]models.Document, error) {
	c.logger.Infof("Buscando documentos da fonte '%s'", source)

	scrollReq := ScrollRequest{
		Limit:       limit,
		WithPayload: true,
		WithVector:  false,
		Filter: map[string]interface{}{
			"must": []map[string]interface{}{
				{
					"key": "source",
					"match": map[string]interface{}{
						"value": source,
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(scrollReq)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar scroll: %w", err)
	}

	url := fmt.Sprintf("%s/collections/%s/points/scroll", c.baseURL, c.collectionName)
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("erro ao fazer scroll: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("erro no scroll: status %d, body: %s", resp.StatusCode, string(body))
	}

	var scrollResponse ScrollResponse
	if err := json.NewDecoder(resp.Body).Decode(&scrollResponse); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %w", err)
	}

	var documents []models.Document
	for _, point := range scrollResponse.Result.Points {
		doc := c.pointToDocument(point)
		documents = append(documents, doc)
	}

	c.logger.Infof("Encontrados %d documentos da fonte '%s'", len(documents), source)
	return documents, nil
}

// GetCollectionInfo retorna informações sobre a coleção
func (c *Client) GetCollectionInfo(ctx context.Context) (*CollectionInfoResponse, error) {
	c.logger.Info("Obtendo informações da coleção")

	url := fmt.Sprintf("%s/collections/%s", c.baseURL, c.collectionName)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter info da coleção: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("erro ao obter info: status %d, body: %s", resp.StatusCode, string(body))
	}

	var collectionInfo CollectionInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&collectionInfo); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %w", err)
	}

	c.logger.Infof("Coleção '%s' tem %d pontos", c.collectionName, collectionInfo.Result.PointsCount)
	return &collectionInfo, nil
}

// pointToDocument converte um PointStruct em Document
func (c *Client) pointToDocument(point PointStruct) models.Document {
	content := ""
	source := ""
	created := time.Time{}
	metadata := make(map[string]string)

	if contentVal, ok := point.Payload["content"]; ok {
		if str, ok := contentVal.(string); ok {
			content = str
		}
	}

	if sourceVal, ok := point.Payload["source"]; ok {
		if str, ok := sourceVal.(string); ok {
			source = str
		}
	}

	if createdVal, ok := point.Payload["created"]; ok {
		if str, ok := createdVal.(string); ok {
			if t, err := time.Parse("2006-01-02T15:04:05Z", str); err == nil {
				created = t
			}
		}
	}

	// Extrair metadata
	for key, value := range point.Payload {
		if len(key) > 9 && key[:9] == "metadata_" {
			metaKey := key[9:]
			if str, ok := value.(string); ok {
				metadata[metaKey] = str
			}
		}
	}

	return models.Document{
		ID:       point.ID,
		Content:  content,
		Source:   source,
		Metadata: metadata,
		Created:  created,
	}
}
