package models

import "time"

const (
	MaxActiveLoansPerReader = 5
	DefaultLoanDurationDays = 14
)

type Loan struct {
	ID         int        `json:"id"`
	BookID     int        `json:"bookId"`
	ReaderID   int        `json:"readerId"`
	BorrowedAt time.Time  `json:"borrowedAt"`
	DueDate    time.Time  `json:"dueDate"`
	ReturnedAt *time.Time `json:"returnedAt"`
	IsReturned bool       `json:"isReturned"`
}
