package services

import (
	"errors"
	"library-management-api/internal/models"
	"library-management-api/internal/storage"
	"strings"
)

var (
	ErrBookNotFound      = errors.New("книга не найдена")
	ErrBookInvalidData   = errors.New("некорректные данные книги")
	ErrBookHasActiveLoan = errors.New("нельзя удалить книгу с активными выдачами")
)

type BookService struct {
	store *storage.MemoryStorage
}

func NewBookService(store *storage.MemoryStorage) *BookService {
	return &BookService{store: store}
}

func ValidateBookInput(b *models.Book) error {
	if strings.TrimSpace(b.Title) == "" || strings.TrimSpace(b.Author) == "" {
		return ErrBookInvalidData
	}
	if b.Year <= 0 || b.TotalCopies < 0 {
		return ErrBookInvalidData
	}
	return nil
}

func (s *BookService) CreateBook(book models.Book) (*models.Book, error) {
	if err := ValidateBookInput(&book); err != nil {
		return nil, err
	}

	s.store.Mu.Lock()
	defer s.store.Mu.Unlock()

	book.ID = s.store.NextBookID
	s.store.NextBookID++
	book.AvailableCopies = book.TotalCopies
	s.store.Books[book.ID] = book

	return &book, nil
}

func (s *BookService) GetAllBooks() []models.Book {
	s.store.Mu.RLock()
	defer s.store.Mu.RUnlock()

	result := make([]models.Book, 0, len(s.store.Books))
	for _, b := range s.store.Books {
		result = append(result, b)
	}
	return result
}

func (s *BookService) GetBookByID(id int) (*models.Book, error) {
	s.store.Mu.RLock()
	defer s.store.Mu.RUnlock()

	book, exists := s.store.Books[id]
	if !exists {
		return nil, ErrBookNotFound
	}
	return &book, nil
}

func (s *BookService) UpdateBook(id int, updated models.Book) (*models.Book, error) {
	if err := ValidateBookInput(&updated); err != nil {
		return nil, err
	}

	s.store.Mu.Lock()
	defer s.store.Mu.Unlock()

	existing, exists := s.store.Books[id]
	if !exists {
		return nil, ErrBookNotFound
	}

	issuedCopies := existing.TotalCopies - existing.AvailableCopies
	if updated.TotalCopies < issuedCopies {
		return nil, ErrBookInvalidData
	}

	existing.Title = updated.Title
	existing.Author = updated.Author
	existing.Year = updated.Year
	existing.Genre = updated.Genre
	existing.TotalCopies = updated.TotalCopies
	existing.AvailableCopies = updated.TotalCopies - issuedCopies

	s.store.Books[id] = existing
	return &existing, nil
}

func (s *BookService) DeleteBook(id int) error {
	s.store.Mu.Lock()
	defer s.store.Mu.Unlock()

	if _, exists := s.store.Books[id]; !exists {
		return ErrBookNotFound
	}

	for _, loan := range s.store.Loans {
		if loan.BookID == id && !loan.IsReturned {
			return ErrBookHasActiveLoan
		}
	}

	delete(s.store.Books, id)
	return nil
}

func (s *BookService) SearchBooks(title, author, genre string, year int) []models.Book {
	s.store.Mu.RLock()
	defer s.store.Mu.RUnlock()

	var result []models.Book
	for _, b := range s.store.Books {
		if title != "" && !strings.Contains(strings.ToLower(b.Title), strings.ToLower(title)) {
			continue
		}
		if author != "" && !strings.Contains(strings.ToLower(b.Author), strings.ToLower(author)) {
			continue
		}
		if genre != "" && !strings.EqualFold(b.Genre, genre) {
			continue
		}
		if year > 0 && b.Year != year {
			continue
		}
		result = append(result, b)
	}
	if result == nil {
		return []models.Book{}
	}
	return result
}
