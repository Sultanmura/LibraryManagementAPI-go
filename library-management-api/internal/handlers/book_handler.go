package handlers

import (
	"encoding/json"
	"errors"
	"library-management-api/internal/models"
	"library-management-api/internal/response"
	"library-management-api/internal/services"
	"net/http"
	"strconv"
)

type BookHandler struct {
	service *services.BookService
}

func NewBookHandler(service *services.BookService) *BookHandler {
	return &BookHandler{service: service}
}

func (h *BookHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	books := h.service.GetAllBooks()
	response.JSON(w, http.StatusOK, books)
}

func (h *BookHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Некорректный ID книги")
		return
	}

	book, err := h.service.GetBookByID(id)
	if err != nil {
		if errors.Is(err, services.ErrBookNotFound) {
			response.Error(w, http.StatusNotFound, "BOOK_NOT_FOUND", err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, book)
}

func (h *BookHandler) Create(w http.ResponseWriter, r *http.Request) {
	var book models.Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "Некорректный формат JSON")
		return
	}

	created, err := h.service.CreateBook(book)
	if err != nil {
		if errors.Is(err, services.ErrBookInvalidData) {
			response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, created)
}

func (h *BookHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Некорректный ID книги")
		return
	}

	var book models.Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "Некорректный формат JSON")
		return
	}

	updated, err := h.service.UpdateBook(id, book)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrBookNotFound):
			response.Error(w, http.StatusNotFound, "BOOK_NOT_FOUND", err.Error())
		case errors.Is(err, services.ErrBookInvalidData):
			response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		}
		return
	}

	response.JSON(w, http.StatusOK, updated)
}

func (h *BookHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Некорректный ID книги")
		return
	}

	if err := h.service.DeleteBook(id); err != nil {
		switch {
		case errors.Is(err, services.ErrBookNotFound):
			response.Error(w, http.StatusNotFound, "BOOK_NOT_FOUND", err.Error())
		case errors.Is(err, services.ErrBookHasActiveLoan):
			response.Error(w, http.StatusConflict, "BOOK_HAS_ACTIVE_LOANS", err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		}
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "Книга успешно удалена"})
}

func (h *BookHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	title := q.Get("title")
	author := q.Get("author")
	genre := q.Get("genre")
	year := 0
	if yearStr := q.Get("year"); yearStr != "" {
		parsedYear, err := strconv.Atoi(yearStr)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "INVALID_YEAR", "Год должен быть числом")
			return
		}
		year = parsedYear
	}

	books := h.service.SearchBooks(title, author, genre, year)
	response.JSON(w, http.StatusOK, books)
}
