package handlers

import (
	"encoding/json"
	"errors"
	"library-management-api/internal/response"
	"library-management-api/internal/services"
	"net/http"
	"strconv"
)

type LoanHandler struct {
	service *services.LoanService
}

func NewLoanHandler(service *services.LoanService) *LoanHandler {
	return &LoanHandler{service: service}
}

type createLoanRequest struct {
	BookID   int `json:"bookId"`
	ReaderID int `json:"readerId"`
}

func (h *LoanHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	loans := h.service.GetAllLoans()
	response.JSON(w, http.StatusOK, loans)
}

func (h *LoanHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Некорректный ID выдачи")
		return
	}

	loan, err := h.service.GetLoanByID(id)
	if err != nil {
		if errors.Is(err, services.ErrLoanNotFound) {
			response.Error(w, http.StatusNotFound, "LOAN_NOT_FOUND", err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, loan)
}

func (h *LoanHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createLoanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "Некорректный формат JSON")
		return
	}

	if req.BookID <= 0 || req.ReaderID <= 0 {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "bookId и readerId должны быть положительными")
		return
	}

	loan, err := h.service.CreateLoan(req.BookID, req.ReaderID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrBookNotFound):
			response.Error(w, http.StatusNotFound, "BOOK_NOT_FOUND", err.Error())
		case errors.Is(err, services.ErrReaderNotFound):
			response.Error(w, http.StatusNotFound, "READER_NOT_FOUND", err.Error())
		case errors.Is(err, services.ErrReaderInactive):
			response.Error(w, http.StatusConflict, "READER_INACTIVE", err.Error())
		case errors.Is(err, services.ErrBookNotAvailable):
			response.Error(w, http.StatusConflict, "BOOK_NOT_AVAILABLE", err.Error())
		case errors.Is(err, services.ErrMaxLoansExceeded):
			response.Error(w, http.StatusConflict, "MAX_LOANS_EXCEEDED", err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		}
		return
	}

	response.JSON(w, http.StatusCreated, loan)
}

func (h *LoanHandler) ReturnBook(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Некорректный ID выдачи")
		return
	}

	loan, err := h.service.ReturnLoan(id)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrLoanNotFound):
			response.Error(w, http.StatusNotFound, "LOAN_NOT_FOUND", err.Error())
		case errors.Is(err, services.ErrLoanAlreadyReturned):
			response.Error(w, http.StatusConflict, "LOAN_ALREADY_RETURNED", err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		}
		return
	}

	response.JSON(w, http.StatusOK, loan)
}

func (h *LoanHandler) GetActive(w http.ResponseWriter, r *http.Request) {
	loans := h.service.GetActiveLoans()
	response.JSON(w, http.StatusOK, loans)
}

func (h *LoanHandler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	stats := h.service.GetStatistics()
	response.JSON(w, http.StatusOK, stats)
}
