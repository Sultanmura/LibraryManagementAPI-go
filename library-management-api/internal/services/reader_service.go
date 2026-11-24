package services

import (
	"errors"
	"library-management-api/internal/models"
	"library-management-api/internal/storage"
	"strings"
)

var (
	ErrReaderNotFound      = errors.New("читатель не найден")
	ErrReaderInvalidData   = errors.New("некорректные данные читателя")
	ErrEmailAlreadyExists  = errors.New("читатель с таким email уже существует")
	ErrReaderHasActiveLoan = errors.New("нельзя удалить читателя с активными выдачами")
)

type ReaderService struct {
	store *storage.MemoryStorage
}

func NewReaderService(store *storage.MemoryStorage) *ReaderService {
	return &ReaderService{store: store}
}

func ValidateReaderInput(r *models.Reader) error {
	if strings.TrimSpace(r.FirstName) == "" || strings.TrimSpace(r.LastName) == "" {
		return ErrReaderInvalidData
	}
	if !strings.Contains(r.Email, "@") || r.Age <= 0 {
		return ErrReaderInvalidData
	}
	return nil
}

func (s *ReaderService) isEmailUniqueLocked(email string, excludeID int) bool {
	for id, r := range s.store.Readers {
		if id != excludeID && strings.EqualFold(r.Email, email) {
			return false
		}
	}
	return true
}

func (s *ReaderService) CreateReader(reader models.Reader) (*models.Reader, error) {
	if err := ValidateReaderInput(&reader); err != nil {
		return nil, err
	}

	s.store.Mu.Lock()
	defer s.store.Mu.Unlock()

	if !s.isEmailUniqueLocked(reader.Email, 0) {
		return nil, ErrEmailAlreadyExists
	}

	reader.ID = s.store.NextReaderID
	s.store.NextReaderID++
	s.store.Readers[reader.ID] = reader

	return &reader, nil
}

func (s *ReaderService) GetAllReaders(nameFilter string) []models.Reader {
	s.store.Mu.RLock()
	defer s.store.Mu.RUnlock()

	result := make([]models.Reader, 0)
	for _, r := range s.store.Readers {
		if nameFilter != "" {
			fullName := strings.ToLower(r.FirstName + " " + r.LastName)
			if !strings.Contains(fullName, strings.ToLower(nameFilter)) {
				continue
			}
		}
		result = append(result, r)
	}
	return result
}

func (s *ReaderService) GetReaderByID(id int) (*models.Reader, error) {
	s.store.Mu.RLock()
	defer s.store.Mu.RUnlock()

	reader, exists := s.store.Readers[id]
	if !exists {
		return nil, ErrReaderNotFound
	}
	return &reader, nil
}

func (s *ReaderService) UpdateReader(id int, updated models.Reader) (*models.Reader, error) {
	if err := ValidateReaderInput(&updated); err != nil {
		return nil, err
	}

	s.store.Mu.Lock()
	defer s.store.Mu.Unlock()

	existing, exists := s.store.Readers[id]
	if !exists {
		return nil, ErrReaderNotFound
	}

	if !s.isEmailUniqueLocked(updated.Email, id) {
		return nil, ErrEmailAlreadyExists
	}

	existing.FirstName = updated.FirstName
	existing.LastName = updated.LastName
	existing.Email = updated.Email
	existing.Age = updated.Age
	existing.IsActive = updated.IsActive

	s.store.Readers[id] = existing
	return &existing, nil
}

func (s *ReaderService) DeleteReader(id int) error {
	s.store.Mu.Lock()
	defer s.store.Mu.Unlock()

	if _, exists := s.store.Readers[id]; !exists {
		return ErrReaderNotFound
	}

	for _, loan := range s.store.Loans {
		if loan.ReaderID == id && !loan.IsReturned {
			return ErrReaderHasActiveLoan
		}
	}

	delete(s.store.Readers, id)
	return nil
}
