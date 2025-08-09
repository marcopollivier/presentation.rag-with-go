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
	for i, doc := range docs {
		contextParts = append(contextParts, fmt.Sprintf("Documento %d (Score: %.2f):\n%s",
			i+1, doc.Score, doc.Document.Content))
	}

	context := strings.Join(contextParts, "\n\n")

	// Prompt para o modelo
	systemPrompt := `Você é um assistente especializado em responder perguntas baseado exclusivamente no contexto fornecido.

INSTRUÇÕES:
1. Use APENAS as informações do contexto fornecido para responder
2. Se a resposta não estiver no contexto, diga claramente que não tem informações suficientes
3. Seja preciso e cite as partes relevantes do contexto
4. Responda em português brasileiro
5. Seja conciso mas completo`

	userPrompt := fmt.Sprintf(`CONTEXTO:
%s

PERGUNTA: %s

RESPOSTA:`, context, query)

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
