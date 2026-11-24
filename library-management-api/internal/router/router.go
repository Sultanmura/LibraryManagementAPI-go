package router

import (
	"library-management-api/internal/handlers"
	"library-management-api/internal/middleware"
	"library-management-api/internal/response"
	"net/http"
	"strings"
)

func NewRouter(
	bookHandler *handlers.BookHandler,
	readerHandler *handlers.ReaderHandler,
	loanHandler *handlers.LoanHandler,
) http.Handler {
	mux := http.NewServeMux()

	// Books (6 endpoints)
	mux.HandleFunc("GET /api/books", bookHandler.GetAll)
	mux.HandleFunc("GET /api/books/search", bookHandler.Search)
	mux.HandleFunc("GET /api/books/{id}", bookHandler.GetByID)
	mux.HandleFunc("POST /api/books", bookHandler.Create)
	mux.HandleFunc("PUT /api/books/{id}", bookHandler.Update)
	mux.HandleFunc("DELETE /api/books/{id}", bookHandler.Delete)

	// Readers (6 endpoints)
	mux.HandleFunc("GET /api/readers", readerHandler.GetAll)
	mux.HandleFunc("GET /api/readers/{id}", readerHandler.GetByID)
	mux.HandleFunc("POST /api/readers", readerHandler.Create)
	mux.HandleFunc("PUT /api/readers/{id}", readerHandler.Update)
	mux.HandleFunc("DELETE /api/readers/{id}", readerHandler.Delete)
	mux.HandleFunc("GET /api/readers/{id}/loans", readerHandler.GetLoansHistory)

	// Loans & Statistics (7 endpoints)
	mux.HandleFunc("GET /api/loans", loanHandler.GetAll)
	mux.HandleFunc("GET /api/loans/active", loanHandler.GetActive)
	mux.HandleFunc("GET /api/loans/{id}", loanHandler.GetByID)
	mux.HandleFunc("POST /api/loans", loanHandler.Create)
	mux.HandleFunc("PUT /api/loans/{id}/return", loanHandler.ReturnBook)
	mux.HandleFunc("GET /api/statistics", loanHandler.GetStatistics)

	// Fallback для 404 и 405
	customWrapper := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, pattern := mux.Handler(r)
		if pattern == "" {
			if isKnownPathWithDifferentMethod(r.URL.Path) {
				response.Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Метод "+r.Method+" не поддерживается для данного маршрута")
				return
			}
			response.Error(w, http.StatusNotFound, "ROUTE_NOT_FOUND", "Маршрут не найден")
			return
		}
		mux.ServeHTTP(w, r)
	})

	return middleware.Logging(customWrapper)
}

func isKnownPathWithDifferentMethod(path string) bool {
	prefixes := []string{"/api/books", "/api/readers", "/api/loans", "/api/statistics"}
	for _, p := range prefixes {
		if path == p || strings.HasPrefix(path, p+"/") {
			return true
		}
	}
	return false
}
