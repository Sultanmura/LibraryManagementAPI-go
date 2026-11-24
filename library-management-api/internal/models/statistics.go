package models

type TopBookStat struct {
	BookID      int    `json:"bookId"`
	Title       string `json:"title"`
	Author      string `json:"author"`
	BorrowCount int    `json:"borrowCount"`
}

type Statistics struct {
	TotalBooks          int           `json:"totalBooks"`
	TotalCopies         int           `json:"totalCopies"`
	AvailableCopies     int           `json:"availableCopies"`
	AvailabilityPercent float64       `json:"availabilityPercent"`
	TotalReaders        int           `json:"totalReaders"`
	ActiveLoans         int           `json:"activeLoans"`
	OverdueLoans        int           `json:"overdueLoans"`
	TopBorrowedBooks    []TopBookStat `json:"topBorrowedBooks"`
	LoansByMonth        [12]int       `json:"loansByMonth"`
}
