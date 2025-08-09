package rag

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/marcopollivier/rag-go-ex01/internal/models"
	"github.com/marcopollivier/rag-go-ex01/internal/openai"
	"github.com/marcopollivier/rag-go-ex01/internal/qdrant"
	"github.com/sirupsen/logrus"
)

type Service struct {
	openaiClient *openai.Client
	qdrantClient *qdrant.Client
	logger       *logrus.Logger
}

// NewService cria um novo serviço RAG
func NewService(openaiClient *openai.Client, qdrantClient *qdrant.Client, logger *logrus.Logger) *Service {
	return &Service{
		openaiClient: openaiClient,
		qdrantClient: qdrantClient,
		logger:       logger,
	}
}

// IndexDocuments indexa uma lista de documentos
func (s *Service) IndexDocuments(ctx context.Context, documents []models.Document) (*models.IndexResponse, error) {
	startTime := time.Now()
	s.logger.Infof("Iniciando indexação de %d documentos", len(documents))

	var failedDocs []string
	indexedCount := 0

	for _, doc := range documents {
		// Gerar ID se não fornecido
		if doc.ID == "" {
			doc.ID = uuid.New().String()
		}

		// Gerar embedding do conteúdo
		embedding, err := s.openaiClient.GenerateEmbedding(ctx, doc.Content)
		if err != nil {
			s.logger.WithError(err).Errorf("Erro ao gerar embedding para documento %s", doc.ID)
			failedDocs = append(failedDocs, doc.ID)
			continue
		}

		// Indexar no Qdrant
		if err := s.qdrantClient.IndexDocument(ctx, doc, embedding); err != nil {
			s.logger.WithError(err).Errorf("Erro ao indexar documento %s", doc.ID)
			failedDocs = append(failedDocs, doc.ID)
			continue
		}

		indexedCount++
	}

	processingTime := time.Since(startTime)
	s.logger.Infof("Indexação concluída: %d sucessos, %d falhas em %v",
		indexedCount, len(failedDocs), processingTime)

	return &models.IndexResponse{
		Success:        len(failedDocs) == 0,
		IndexedCount:   indexedCount,
		FailedDocs:     failedDocs,
		ProcessingTime: processingTime.String(),
	}, nil
}

// Query executa uma consulta RAG
func (s *Service) Query(ctx context.Context, req models.QueryRequest) (*models.QueryResponse, error) {
	startTime := time.Now()
	s.logger.Infof("Executando query RAG: %s", req.Query)

	// Configurar valores padrão
	if req.TopK == 0 {
		req.TopK = 5
	}
	if req.Threshold == 0 {
		req.Threshold = 0.7
	}

	// Gerar embedding da query
	queryEmbedding, err := s.openaiClient.GenerateEmbedding(ctx, req.Query)
	if err != nil {
		s.logger.WithError(err).Error("Erro ao gerar embedding da query")
		return nil, fmt.Errorf("erro ao gerar embedding da query: %w", err)
	}

	// Buscar documentos similares
	relevantDocs, err := s.qdrantClient.SearchSimilar(ctx, queryEmbedding, req.TopK, req.Threshold)
	if err != nil {
		s.logger.WithError(err).Error("Erro ao buscar documentos similares")
		return nil, fmt.Errorf("erro ao buscar documentos similares: %w", err)
	}

	// Gerar resposta usando OpenAI
	answer, err := s.openaiClient.GenerateAnswer(ctx, req.Query, relevantDocs)
	if err != nil {
		s.logger.WithError(err).Error("Erro ao gerar resposta")
		return nil, fmt.Errorf("erro ao gerar resposta: %w", err)
	}

	processingTime := time.Since(startTime)
	s.logger.Infof("Query processada em %v com %d documentos relevantes",
		processingTime, len(relevantDocs))

	return &models.QueryResponse{
		Answer:           answer,
		RelevantDocs:     relevantDocs,
		ProcessingTimeMs: processingTime.Milliseconds(),
	}, nil
}

// IndexTextFiles indexa arquivos de texto de uma pasta
func (s *Service) IndexTextFiles(ctx context.Context, folderPath string) (*models.IndexResponse, error) {
	s.logger.Infof("Indexando arquivos de texto da pasta: %s", folderPath)

	var documents []models.Document

	// Verificar se a pasta existe
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		s.logger.Warnf("Pasta %s não existe, criando documentos de exemplo", folderPath)
		// Se a pasta não existir, criar documentos de exemplo para demo
		documents = s.createSampleDocuments()
	} else {
		// Ler arquivos da pasta
		files, err := ioutil.ReadDir(folderPath)
		if err != nil {
			return nil, fmt.Errorf("erro ao ler pasta %s: %w", folderPath, err)
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			// Processar apenas arquivos .txt
			if !strings.HasSuffix(strings.ToLower(file.Name()), ".txt") {
				continue
			}

			filePath := filepath.Join(folderPath, file.Name())
			content, err := ioutil.ReadFile(filePath)
			if err != nil {
				s.logger.WithError(err).Warnf("Erro ao ler arquivo %s", filePath)
				continue
			}

			// Criar documento
			doc := models.Document{
				ID:      uuid.New().String(),
				Content: string(content),
				Source:  file.Name(),
				Metadata: map[string]string{
					"file_path": filePath,
					"file_size": fmt.Sprintf("%d", len(content)),
					"language":  "portuguese",
				},
				Created: time.Now(),
			}

			documents = append(documents, doc)
			s.logger.Infof("Arquivo %s lido com %d caracteres", file.Name(), len(content))
		}

		if len(documents) == 0 {
			s.logger.Warn("Nenhum arquivo .txt encontrado, criando documentos de exemplo")
			documents = s.createSampleDocuments()
		}
	}

	return s.IndexDocuments(ctx, documents)
}

// createSampleDocuments cria documentos de exemplo para demo
func (s *Service) createSampleDocuments() []models.Document {
	return []models.Document{
		{
			ID:      uuid.New().String(),
			Content: "Go é uma linguagem de programação desenvolvida pelo Google. É conhecida por sua simplicidade, performance e excelente suporte a concorrência através de goroutines e channels. A linguagem foi criada em 2007 por Robert Griesemer, Rob Pike e Ken Thompson na Google. O objetivo era criar uma linguagem que combinasse a facilidade de programação de uma linguagem interpretada dinamicamente com a eficiência e segurança de uma linguagem compilada estaticamente. Principais características do Go: compilação rápida, garbage collection eficiente, sistema de tipos forte, concorrência nativa, sintaxe simples e limpa, cross-platform, excelente para desenvolvimento de microserviços e APIs.",
			Source:  "golang_intro.txt",
			Metadata: map[string]string{
				"category": "programming",
				"language": "portuguese",
			},
			Created: time.Now(),
		},
		{
			ID:      uuid.New().String(),
			Content: "RAG (Retrieval Augmented Generation) é uma técnica avançada de inteligência artificial que combina busca de informações com geração de texto. O conceito funciona em duas etapas principais: 1. Retrieval (Recuperação): Busca documentos ou trechos de texto relevantes em uma base de conhecimento 2. Generation (Geração): Usa um modelo de linguagem para gerar uma resposta baseada no contexto recuperado. Vantagens do RAG: reduz alucinações do modelo, permite uso de conhecimento específico e atualizado, não requer retreinamento do modelo, transparência sobre as fontes de informação, escalabilidade para grandes bases de conhecimento. O RAG é especialmente útil para: sistemas de perguntas e respostas, assistentes virtuais especializados, chatbots corporativos, análise de documentos, suporte técnico automatizado.",
			Source:  "rag_explanation.txt",
			Metadata: map[string]string{
				"category": "ai",
				"language": "portuguese",
			},
			Created: time.Now(),
		},
		{
			ID:      uuid.New().String(),
			Content: "Qdrant é um banco de dados vetorial de código aberto, otimizado para aplicações de machine learning e busca semântica. Oferece APIs RESTful e gRPC para operações de inserção, busca e filtragem de vetores. Suporta filtros avançados, clustering de vetores, e indexação eficiente usando algoritmo HNSW. É ideal para aplicações de recomendação, busca semântica, detecção de similaridade e sistemas RAG. Qdrant pode ser executado como um serviço standalone ou integrado em aplicações através de suas bibliotecas cliente.",
			Source:  "qdrant_info.txt",
			Metadata: map[string]string{
				"category": "database",
				"language": "portuguese",
			},
			Created: time.Now(),
		},
	}
}

// GetStats retorna estatísticas do sistema RAG
func (s *Service) GetStats(ctx context.Context) (map[string]interface{}, error) {
	s.logger.Debug("Obtendo estatísticas do sistema")

	collectionInfo, err := s.qdrantClient.GetCollectionInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter informações da coleção: %w", err)
	}

	return collectionInfo, nil
}
