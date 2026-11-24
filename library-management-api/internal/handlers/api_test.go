package handlers_test

import (
	"bytes"
	"encoding/json"
	"library-management-api/internal/handlers"
	"library-management-api/internal/models"
	"library-management-api/internal/response"
	"library-management-api/internal/router"
	"library-management-api/internal/services"
	"library-management-api/internal/storage"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupTestServer() (http.Handler, *services.BookService, *services.ReaderService, *services.LoanService) {
	store := storage.NewMemoryStorage()
	bs := services.NewBookService(store)
	rs := services.NewReaderService(store)
	ls := services.NewLoanService(store)

	bh := handlers.NewBookHandler(bs)
	rh := handlers.NewReaderHandler(rs, ls)
	lh := handlers.NewLoanHandler(ls)

	return router.NewRouter(bh, rh, lh), bs, rs, ls
}

func TestCreateAndGetBookHTTP(t *testing.T) {
	srv, _, _, _ := setupTestServer()

	body := `{"title":"Refactoring","author":"Martin Fowler","year":2018,"genre":"IT","totalCopies":4}`
	req := httptest.NewRequest(http.MethodPost, "/api/books", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("ожидался статус 201, получен %d", rec.Code)
	}

	reqGet := httptest.NewRequest(http.MethodGet, "/api/books/1", nil)
	recGet := httptest.NewRecorder()
	srv.ServeHTTP(recGet, reqGet)

	if recGet.Code != http.StatusOK {
		t.Fatalf("ожидался статус 200, получен %d", recGet.Code)
	}
}

func TestDuplicateEmailConflictHTTP(t *testing.T) {
	srv, _, _, _ := setupTestServer()

	body := `{"firstName":"Иван","lastName":"Иванов","email":"ivan@test.kz","age":25,"isActive":true}`
	req1 := httptest.NewRequest(http.MethodPost, "/api/readers", bytes.NewBufferString(body))
	rec1 := httptest.NewRecorder()
	srv.ServeHTTP(rec1, req1)

	if rec1.Code != http.StatusCreated {
		t.Fatalf("ожидался статус 201, получен %d", rec1.Code)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/api/readers", bytes.NewBufferString(body))
	rec2 := httptest.NewRecorder()
	srv.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusConflict {
		t.Fatalf("ожидался статус 409, получен %d", rec2.Code)
	}
}

func TestLoanAndReturnFlowHTTP(t *testing.T) {
	srv, bs, rs, _ := setupTestServer()

	book, _ := bs.CreateBook(models.Book{Title: "Go", Author: "Pike", Year: 2020, Genre: "IT", TotalCopies: 1})
	reader, _ := rs.CreateReader(models.Reader{FirstName: "Айгерим", LastName: "Ким", Email: "aig@test.kz", Age: 22, IsActive: true})

	loanPayload, _ := json.Marshal(map[string]int{"bookId": book.ID, "readerId": reader.ID})
	reqLoan := httptest.NewRequest(http.MethodPost, "/api/loans", bytes.NewReader(loanPayload))
	recLoan := httptest.NewRecorder()
	srv.ServeHTTP(recLoan, reqLoan)

	if recLoan.Code != http.StatusCreated {
		t.Fatalf("ожидался статус 201 при выдаче, получен %d", recLoan.Code)
	}

	reqLoan2 := httptest.NewRequest(http.MethodPost, "/api/loans", bytes.NewReader(loanPayload))
	recLoan2 := httptest.NewRecorder()
	srv.ServeHTTP(recLoan2, reqLoan2)

	if recLoan2.Code != http.StatusConflict {
		t.Fatalf("ожидался статус 409 при отсутствии копий, получен %d", recLoan2.Code)
	}

	reqReturn := httptest.NewRequest(http.MethodPut, "/api/loans/1/return", nil)
	recReturn := httptest.NewRecorder()
	srv.ServeHTTP(recReturn, reqReturn)

	if recReturn.Code != http.StatusOK {
		t.Fatalf("ожидался статус 200 при возврате, получен %d", recReturn.Code)
	}

	reqReturn2 := httptest.NewRequest(http.MethodPut, "/api/loans/1/return", nil)
	recReturn2 := httptest.NewRecorder()
	srv.ServeHTTP(recReturn2, reqReturn2)

	if recReturn2.Code != http.StatusConflict {
		t.Fatalf("ожидался статус 409 при повторном возврате, получен %d", recReturn2.Code)
	}
}

func TestMaxLoansLimitService(t *testing.T) {
	_, bs, rs, ls := setupTestServer()

	book, _ := bs.CreateBook(models.Book{Title: "Algorithms", Author: "Cormen", Year: 2009, Genre: "CS", TotalCopies: 10})
	reader, _ := rs.CreateReader(models.Reader{FirstName: "Данияр", LastName: "Алиев", Email: "dan@test.kz", Age: 20, IsActive: true})

	for i := 0; i < 5; i++ {
		if _, err := ls.CreateLoan(book.ID, reader.ID); err != nil {
			t.Fatalf("ошибка на выдаче #%d: %v", i+1, err)
		}
	}

	if _, err := ls.CreateLoan(book.ID, reader.ID); err != services.ErrMaxLoansExceeded {
		t.Fatalf("ожидалась ошибка ErrMaxLoansExceeded на 6-й выдаче, получена: %v", err)
	}
}

func TestDeleteBookWithActiveLoanConflict(t *testing.T) {
	_, bs, rs, ls := setupTestServer()

	book, _ := bs.CreateBook(models.Book{Title: "DDD", Author: "Evans", Year: 2003, Genre: "IT", TotalCopies: 2})
	reader, _ := rs.CreateReader(models.Reader{FirstName: "Ерлан", LastName: "Нуров", Email: "erl@test.kz", Age: 24, IsActive: true})
	_, _ = ls.CreateLoan(book.ID, reader.ID)

	if err := bs.DeleteBook(book.ID); err != services.ErrBookHasActiveLoan {
		t.Fatalf("ожидалась ошибка ErrBookHasActiveLoan, получена: %v", err)
	}
}

func TestStatisticsCalculation(t *testing.T) {
	srv, bs, rs, ls := setupTestServer()

	b1, _ := bs.CreateBook(models.Book{Title: "Book 1", Author: "A", Year: 2021, Genre: "IT", TotalCopies: 4})
	r1, _ := rs.CreateReader(models.Reader{FirstName: "Тест", LastName: "Тестов", Email: "t@test.kz", Age: 30, IsActive: true})
	_, _ = ls.CreateLoan(b1.ID, r1.ID)

	req := httptest.NewRequest(http.MethodGet, "/api/statistics", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	var resp response.APIResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("ошибка декодирования ответа: %v", err)
	}
	if !resp.IsSuccess || resp.StatusCode != 200 {
		t.Fatalf("некорректный ответ статистики: %+v", resp)
	}
}
