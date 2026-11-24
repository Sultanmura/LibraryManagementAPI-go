package storage

import (
	"library-management-api/internal/models"
	"sync"
)

type MemoryStorage struct {
	Mu           sync.RWMutex
	Books        map[int]models.Book
	Readers      map[int]models.Reader
	Loans        map[int]models.Loan
	NextBookID   int
	NextReaderID int
	NextLoanID   int
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		Books:        make(map[int]models.Book),
		Readers:      make(map[int]models.Reader),
		Loans:        make(map[int]models.Loan),
		NextBookID:   1,
		NextReaderID: 1,
		NextLoanID:   1,
	}
}
