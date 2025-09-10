package openai

import (
	"context"
	"fmt"
	"strings"

	"github.com/marcopollivier/rag-go-ex01/internal/models"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

type Client struct {
	client *openai.Client
	model  string
	logger *logrus.Logger
}

// NewClient cria um novo cliente OpenAI
func NewClient(apiKey string, logger *logrus.Logger) *Client {
	return &Client{
		client: openai.NewClient(apiKey),
		model:  openai.GPT3Dot5Turbo,
		logger: logger,
	}
}

// GenerateEmbedding gera embedding para um texto
func (c *Client) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	c.logger.Debugf("Gerando embedding para texto de %d caracteres", len(text))

	req := openai.EmbeddingRequest{
		Input: []string{text},
		Model: openai.AdaEmbeddingV2,
	}

	resp, err := c.client.CreateEmbeddings(ctx, req)
	if err != nil {
		c.logger.WithError(err).Error("Erro ao gerar embedding")
		return nil, fmt.Errorf("erro ao gerar embedding: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("nenhum embedding retornado")
	}

	return resp.Data[0].Embedding, nil
}

// GenerateAnswer gera uma resposta baseada no contexto e pergunta
func (c *Client) GenerateAnswer(ctx context.Context, query string, docs []models.RelevantDocument) (string, error) {
	c.logger.Debugf("Gerando resposta para query: %s com %d documentos", query, len(docs))

	// Construir contexto a partir dos documentos relevantes
	var contextParts []string
	hasRelevantDocs := false

	for i, doc := range docs {
		if doc.Score >= 0.75 {
			hasRelevantDocs = true
			contextParts = append(contextParts, fmt.Sprintf("Documento %d (Score: %.2f):\n%s",
				i+1, doc.Score, doc.Document.Content))
		}
	}

	var userPrompt string
	if hasRelevantDocs {
		context := strings.Join(contextParts, "\n\n")
		userPrompt = fmt.Sprintf(`CONTEXTO RELEVANTE:
%s

PERGUNTA: %s

RESPOSTA (baseada no contexto):`, context, query)
	} else {
		userPrompt = fmt.Sprintf(`PERGUNTA: %s

Responda de forma direta e objetiva, sem mencionar fontes ou falta de informações:`, query)
	}

	systemPrompt := `Você é um assistente inteligente que responde perguntas de forma útil e objetiva.

INSTRUÇÕES:
1. Responda de forma direta e objetiva, sem mencionar fontes ou contexto
2. Use informações do contexto quando fornecidas e relevantes
3. Use conhecimento geral quando o contexto não for relevante para a pergunta
4. Responda em português brasileiro
5. Seja preciso, útil e completo
6. Vá direto ao ponto, sem preâmbulos ou explicações sobre as fontes de informação`

	req := openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: userPrompt,
			},
		},
		MaxTokens:   1000,
		Temperature: 0.1,
	}

	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		c.logger.WithError(err).Error("Erro ao gerar resposta")
		return "", fmt.Errorf("erro ao gerar resposta: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("nenhuma resposta gerada")
	}

	answer := resp.Choices[0].Message.Content
	c.logger.Debugf("Resposta gerada com %d caracteres", len(answer))

	return answer, nil
}
