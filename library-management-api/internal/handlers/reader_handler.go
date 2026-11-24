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

type ReaderHandler struct {
	service     *services.ReaderService
	loanService *services.LoanService
}

func NewReaderHandler(service *services.ReaderService, loanService *services.LoanService) *ReaderHandler {
	return &ReaderHandler{service: service, loanService: loanService}
}

func (h *ReaderHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	readers := h.service.GetAllReaders(name)
	response.JSON(w, http.StatusOK, readers)
}

func (h *ReaderHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Некорректный ID читателя")
		return
	}

	reader, err := h.service.GetReaderByID(id)
	if err != nil {
		if errors.Is(err, services.ErrReaderNotFound) {
			response.Error(w, http.StatusNotFound, "READER_NOT_FOUND", err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, reader)
}

func (h *ReaderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var reader models.Reader
	if err := json.NewDecoder(r.Body).Decode(&reader); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "Некорректный формат JSON")
		return
	}

	created, err := h.service.CreateReader(reader)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrReaderInvalidData):
			response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		case errors.Is(err, services.ErrEmailAlreadyExists):
			response.Error(w, http.StatusConflict, "EMAIL_ALREADY_EXISTS", err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		}
		return
	}

	response.JSON(w, http.StatusCreated, created)
}

func (h *ReaderHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Некорректный ID читателя")
		return
	}

	var reader models.Reader
	if err := json.NewDecoder(r.Body).Decode(&reader); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "Некорректный формат JSON")
		return
	}

	updated, err := h.service.UpdateReader(id, reader)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrReaderNotFound):
			response.Error(w, http.StatusNotFound, "READER_NOT_FOUND", err.Error())
		case errors.Is(err, services.ErrReaderInvalidData):
			response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		case errors.Is(err, services.ErrEmailAlreadyExists):
			response.Error(w, http.StatusConflict, "EMAIL_ALREADY_EXISTS", err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		}
		return
	}

	response.JSON(w, http.StatusOK, updated)
}

func (h *ReaderHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Некорректный ID читателя")
		return
	}

	if err := h.service.DeleteReader(id); err != nil {
		switch {
		case errors.Is(err, services.ErrReaderNotFound):
			response.Error(w, http.StatusNotFound, "READER_NOT_FOUND", err.Error())
		case errors.Is(err, services.ErrReaderHasActiveLoan):
			response.Error(w, http.StatusConflict, "READER_HAS_ACTIVE_LOANS", err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		}
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "Читатель успешно удалён"})
}

func (h *ReaderHandler) GetLoansHistory(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Некорректный ID читателя")
		return
	}

	loans, err := h.loanService.GetLoansByReaderID(id)
	if err != nil {
		if errors.Is(err, services.ErrReaderNotFound) {
			response.Error(w, http.StatusNotFound, "READER_NOT_FOUND", err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, loans)
}
