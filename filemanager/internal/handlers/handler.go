package handlers

import (
	"context"
	"sync"

	"github.com/edgarcoime/Cthulhu-common/pkg/rabbitmq/manager"
	"github.com/edgarcoime/Cthulhu-filemanager/internal/service"
)

// chunkStorage holds chunks for reassembly
type chunkStorage struct {
	mu          sync.Mutex
	chunks      map[string]map[int][]byte // transactionID -> chunkIndex -> data
	chunkCounts map[string]int            // transactionID -> total chunks expected
	metadata    map[string]*chunkMetadata // transactionID -> metadata
}

type chunkMetadata struct {
	filename  string
	storageID string
	totalSize int64
}

// Handler holds dependencies for message handlers
type Handler struct {
	service      service.Service
	manager      *manager.Manager
	ctx          context.Context
	chunkStorage *chunkStorage
}

// NewHandler creates a new handler instance
func NewHandler(service service.Service, manager *manager.Manager, ctx context.Context) *Handler {
	return &Handler{
		service: service,
		manager: manager,
		ctx:     ctx,
		chunkStorage: &chunkStorage{
			chunks:      make(map[string]map[int][]byte),
			chunkCounts: make(map[string]int),
			metadata:    make(map[string]*chunkMetadata),
		},
	}
}
