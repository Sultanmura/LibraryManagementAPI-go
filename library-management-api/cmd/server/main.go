package main

import (
	"library-management-api/internal/handlers"
	"library-management-api/internal/models"
	"library-management-api/internal/router"
	"library-management-api/internal/services"
	"library-management-api/internal/storage"
	"log"
	"net/http"
)

func main() {
	store := storage.NewMemoryStorage()

	bookService := services.NewBookService(store)
	readerService := services.NewReaderService(store)
	loanService := services.NewLoanService(store)

	seedInitialData(bookService, readerService)

	bookHandler := handlers.NewBookHandler(bookService)
	readerHandler := handlers.NewReaderHandler(readerService, loanService)
	loanHandler := handlers.NewLoanHandler(loanService)

	handler := router.NewRouter(bookHandler, readerHandler, loanHandler)

	log.Println("Library Management API запущен на http://localhost:8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}

func seedInitialData(bs *services.BookService, rs *services.ReaderService) {
	_, _ = bs.CreateBook(models.Book{
		Title:       "Clean Code",
		Author:      "Robert C. Martin",
		Year:        2008,
		Genre:       "IT",
		TotalCopies: 3,
	})
	_, _ = bs.CreateBook(models.Book{
		Title:       "The Go Programming Language",
		Author:      "Alan Donovan",
		Year:        2015,
		Genre:       "IT",
		TotalCopies: 2,
	})
	_, _ = rs.CreateReader(models.Reader{
		FirstName: "Алихан",
		LastName:  "Сериков",
		Email:     "alikhan@example.com",
		Age:       21,
		IsActive:  true,
	})
}
