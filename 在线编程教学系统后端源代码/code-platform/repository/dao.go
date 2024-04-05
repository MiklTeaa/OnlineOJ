package repository

import (
	"code-platform/storage"
)

// Dao dao层
type Dao struct {
	Storage *storage.Storage
}

// NewDao 返回dao层
func NewDao() *Dao {
	storage := storage.NewStorage()
	// init Storage
	Storage = storage
	return &Dao{Storage: storage}
}

// Storage : For signal to Close
var Storage *storage.Storage
