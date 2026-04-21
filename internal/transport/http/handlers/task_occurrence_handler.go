package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	taskdomain "example.com/taskservice/internal/domain/task"
	taskoccurrencedomain "example.com/taskservice/internal/domain/task_occurrence"
	taskoccurrenceusecase "example.com/taskservice/internal/usecase/task"
)

type TaskOccurrenceHandler struct {
	usecase *taskoccurrenceusecase.TaskOccurrenceService
}

func NewTaskOccurrenceHandler(usecase *taskoccurrenceusecase.TaskOccurrenceService) *TaskOccurrenceHandler {
	return &TaskOccurrenceHandler{usecase: usecase}
}

func (h *TaskOccurrenceHandler) GetByDate(w http.ResponseWriter, r *http.Request) {
	rawDate := r.URL.Query().Get("date")
	if rawDate == "" {
		writeError(w, http.StatusBadRequest, errors.New("date query param is required"))
		return
	}

	date, err := time.Parse("2006-01-02", rawDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid date format, expected YYYY-MM-DD"))
		return
	}

	occurrences, err := h.usecase.GetByDate(r.Context(), date)
	if err != nil {
		writeOccurrenceUsecaseError(w, err)
		return
	}

	response := make([]taskOccurrenceDTO, 0, len(occurrences))
	for i := range occurrences {
		response = append(response, newTaskOccurrenceDTO(&occurrences[i]))
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *TaskOccurrenceHandler) GetByTaskID(w http.ResponseWriter, r *http.Request) {
	taskID, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	occurrences, err := h.usecase.GetByTaskID(r.Context(), taskID)
	if err != nil {
		writeOccurrenceUsecaseError(w, err)
		return
	}

	response := make([]taskOccurrenceDTO, 0, len(occurrences))
	for i := range occurrences {
		response = append(response, newTaskOccurrenceDTO(&occurrences[i]))
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *TaskOccurrenceHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := getOccurrenceIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var req taskOccurrenceStatusMutationDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.usecase.UpdateStatus(r.Context(), id, req.Status); err != nil {
		writeOccurrenceUsecaseError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func getOccurrenceIDFromRequest(r *http.Request) (int64, error) {
	rawID := mux.Vars(r)["id"]
	if rawID == "" {
		return 0, errors.New("missing occurrence id")
	}

	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid occurrence id")
	}

	return id, nil
}

func writeOccurrenceUsecaseError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, taskdomain.ErrNotFound):
		writeError(w, http.StatusNotFound, err)
	case errors.Is(err, taskoccurrencedomain.ErrNotFound):
		writeError(w, http.StatusNotFound, err)
	case errors.Is(err, taskoccurrenceusecase.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err)
	default:
		writeError(w, http.StatusInternalServerError, err)
	}
}
