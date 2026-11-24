package services

import (
	"errors"
	"library-management-api/internal/models"
	"library-management-api/internal/storage"
	"sort"
	"time"
)

var (
	ErrLoanNotFound        = errors.New("запись о выдаче не найдена")
	ErrBookNotAvailable    = errors.New("нет доступных экземпляров книги")
	ErrReaderInactive      = errors.New("читатель неактивен")
	ErrMaxLoansExceeded    = errors.New("превышен лимит активных выдач (максимум 5)")
	ErrLoanAlreadyReturned = errors.New("книга уже возвращена")
)

type LoanService struct {
	store *storage.MemoryStorage
}

func NewLoanService(store *storage.MemoryStorage) *LoanService {
	return &LoanService{store: store}
}

func (s *LoanService) countActiveLoansByReaderLocked(readerID int) int {
	count := 0
	for _, l := range s.store.Loans {
		if l.ReaderID == readerID && !l.IsReturned {
			count++
		}
	}
	return count
}

func decrementAvailableCopies(book *models.Book) {
	book.AvailableCopies--
}

func incrementAvailableCopies(book *models.Book) {
	book.AvailableCopies++
}

func markLoanReturned(loan *models.Loan, returnTime time.Time) {
	loan.IsReturned = true
	loan.ReturnedAt = &returnTime
}

func (s *LoanService) CreateLoan(bookID, readerID int) (*models.Loan, error) {
	s.store.Mu.Lock()
	defer s.store.Mu.Unlock()

	book, bookExists := s.store.Books[bookID]
	if !bookExists {
		return nil, ErrBookNotFound
	}

	reader, readerExists := s.store.Readers[readerID]
	if !readerExists {
		return nil, ErrReaderNotFound
	}

	if !reader.IsActive {
		return nil, ErrReaderInactive
	}

	if book.AvailableCopies <= 0 {
		return nil, ErrBookNotAvailable
	}

	if s.countActiveLoansByReaderLocked(readerID) >= models.MaxActiveLoansPerReader {
		return nil, ErrMaxLoansExceeded
	}

	decrementAvailableCopies(&book)
	s.store.Books[bookID] = book

	now := time.Now()
	loan := models.Loan{
		ID:         s.store.NextLoanID,
		BookID:     bookID,
		ReaderID:   readerID,
		BorrowedAt: now,
		DueDate:    now.AddDate(0, 0, models.DefaultLoanDurationDays),
		ReturnedAt: nil,
		IsReturned: false,
	}
	s.store.NextLoanID++
	s.store.Loans[loan.ID] = loan

	return &loan, nil
}

func (s *LoanService) ReturnLoan(loanID int) (*models.Loan, error) {
	s.store.Mu.Lock()
	defer s.store.Mu.Unlock()

	loan, exists := s.store.Loans[loanID]
	if !exists {
		return nil, ErrLoanNotFound
	}

	if loan.IsReturned {
		return nil, ErrLoanAlreadyReturned
	}

	markLoanReturned(&loan, time.Now())
	s.store.Loans[loanID] = loan

	if book, bookExists := s.store.Books[loan.BookID]; bookExists {
		incrementAvailableCopies(&book)
		s.store.Books[loan.BookID] = book
	}

	return &loan, nil
}

func (s *LoanService) GetAllLoans() []models.Loan {
	s.store.Mu.RLock()
	defer s.store.Mu.RUnlock()

	result := make([]models.Loan, 0, len(s.store.Loans))
	for _, l := range s.store.Loans {
		result = append(result, l)
	}
	return result
}

func (s *LoanService) GetLoanByID(id int) (*models.Loan, error) {
	s.store.Mu.RLock()
	defer s.store.Mu.RUnlock()

	loan, exists := s.store.Loans[id]
	if !exists {
		return nil, ErrLoanNotFound
	}
	return &loan, nil
}

func (s *LoanService) GetActiveLoans() []models.Loan {
	s.store.Mu.RLock()
	defer s.store.Mu.RUnlock()

	result := make([]models.Loan, 0)
	for _, l := range s.store.Loans {
		if !l.IsReturned {
			result = append(result, l)
		}
	}
	return result
}

func (s *LoanService) GetLoansByReaderID(readerID int) ([]models.Loan, error) {
	s.store.Mu.RLock()
	defer s.store.Mu.RUnlock()

	if _, exists := s.store.Readers[readerID]; !exists {
		return nil, ErrReaderNotFound
	}

	result := make([]models.Loan, 0)
	for _, l := range s.store.Loans {
		if l.ReaderID == readerID {
			result = append(result, l)
		}
	}
	return result, nil
}

func calculateAvailabilityPercent(total, available int) float64 {
	if total == 0 {
		return 0.0
	}
	return (float64(available) / float64(total)) * 100.0
}

func (s *LoanService) GetStatistics() models.Statistics {
	s.store.Mu.RLock()
	defer s.store.Mu.RUnlock()

	var stats models.Statistics
	stats.TotalBooks = len(s.store.Books)
	stats.TotalReaders = len(s.store.Readers)

	for _, b := range s.store.Books {
		stats.TotalCopies += b.TotalCopies
		stats.AvailableCopies += b.AvailableCopies
	}
	stats.AvailabilityPercent = calculateAvailabilityPercent(stats.TotalCopies, stats.AvailableCopies)

	now := time.Now()
	borrowCounts := make(map[int]int)
	var loansByMonth [12]int

	for _, l := range s.store.Loans {
		if !l.IsReturned {
			stats.ActiveLoans++
			if now.After(l.DueDate) {
				stats.OverdueLoans++
			}
		}
		borrowCounts[l.BookID]++
		monthIdx := int(l.BorrowedAt.Month()) - 1
		if monthIdx >= 0 && monthIdx < 12 {
			loansByMonth[monthIdx]++
		}
	}
	stats.LoansByMonth = loansByMonth

	topList := make([]models.TopBookStat, 0, len(borrowCounts))
	for bookID, count := range borrowCounts {
		book, exists := s.store.Books[bookID]
		title := "Удалённая книга"
		author := "Неизвестен"
		if exists {
			title = book.Title
			author = book.Author
		}
		topList = append(topList, models.TopBookStat{
			BookID:      bookID,
			Title:       title,
			Author:      author,
			BorrowCount: count,
		})
	}

	sort.Slice(topList, func(i, j int) bool {
		return topList[i].BorrowCount > topList[j].BorrowCount
	})

	if len(topList) > 3 {
		topList = topList[:3]
	}
	stats.TopBorrowedBooks = topList

	return stats
}
