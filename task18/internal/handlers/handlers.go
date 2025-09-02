package handlers

import (
	"calendar/internal/calendar"
	"calendar/internal/event"
	"calendar/internal/event/repository"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	_ "calendar/docs"
)

type Handlers struct {
	serviceCalendar *calendar.ServiceCalendar
	log             *slog.Logger
}

func NewHandlers(serviceCalendar *calendar.ServiceCalendar, logger *slog.Logger) *Handlers {
	return &Handlers{
		serviceCalendar: serviceCalendar,
		log:             logger,
	}
}

// CreateEvent создает новое событие
// @Summary Создать новое событие
// @Description Создает новое событие в календаре пользователя
// @Tags events
// @Accept json
// @Produce json
// @Param event body repository.CreateEventRequest true "Данные события" SchemaExample({"user_id": 1, "date": "YYYY-MM-DD", "title": "example string"})
// @Success 200 {object} repository.SuccessResponse{result=repository.Event}
// @Failure 400 {object} repository.ErrorResponse
// @Failure 503 {object} repository.ErrorResponse
// @Router /create_event [post]
func (h *Handlers) CreateEvent(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var req repository.CreateEventRequest
	if err := dec.Decode(&req); err != nil {
		sendError(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if err := event.ValidateCreateRequest(req.UserID, req.Date, req.Title); err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	date, err := event.ParseAndValidateDate(req.Date)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdEvent, err := h.serviceCalendar.CreateEvent(req.UserID, date, req.Title)
	if err != nil {
		sendError(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	h.log.Debug("Event created in handle",
		"event_id", createdEvent.ID,
		"user_id", createdEvent.UserID,
		"title", createdEvent.Title,
	)

	sendResponse(w, createdEvent, http.StatusOK)
}

// UpdateEvent обновляет существующее событие
// @Summary Обновить событие
// @Description Обновляет существующее событие в календаре пользователя
// @Tags events
// @Accept json
// @Produce json
// @Param event body repository.UpdateEventRequest true "Данные для обновления события" SchemaExample({"event_id": 1, "user_id": 1, "date": "YYYY-MM-DD", "title": "example string"})
// @Success 200 {object} repository.SuccessResponse{result=repository.Event}
// @Failure 400 {object} repository.ErrorResponse
// @Failure 503 {object} repository.ErrorResponse
// @Router /update_event [post]
func (h *Handlers) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	var req repository.UpdateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "invalid JSON format", http.StatusBadRequest)
		return
	}

	if err := event.ValidateUpdateRequest(req.EventID, req.UserID, req.Date, req.Title); err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	date, err := event.ParseAndValidateDate(req.Date)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	updatedEvent, err := h.serviceCalendar.UpdateEvent(req.EventID, req.UserID, date, req.Title)
	if err != nil {
		sendError(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	h.log.Debug("Event updated in handle",
		"event_id", updatedEvent.ID,
		"user_id", updatedEvent.UserID,
		"title", updatedEvent.Title,
	)

	sendResponse(w, updatedEvent, http.StatusOK)
}

// DeleteEvent удаляет событие
// @Summary Удалить событие
// @Description Удаляет событие из календаря пользователя
// @Tags events
// @Accept json
// @Produce json
// @Param event body repository.DeleteEventRequest true "Данные для удаления события" SchemaExample({"event_id": 1, "user_id": 1})
// @Success 200 {object} repository.SuccessResponse{result=object}
// @Failure 400 {object} repository.ErrorResponse
// @Failure 503 {object} repository.ErrorResponse
// @Router /delete_event [post]
func (h *Handlers) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	var req repository.DeleteEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "invalid JSON format", http.StatusBadRequest)
		return
	}

	if err := event.ValidateDeleteRequest(req.EventID, req.UserID); err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.serviceCalendar.DeleteEvent(req.EventID, req.UserID)
	if err != nil {
		sendError(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	h.log.Debug("Event deleted in handle",
		"event_id", req.EventID,
		"user_id", req.UserID,
	)

	sendResponse(w, map[string]string{"result": "event deleted successfully"}, http.StatusOK)
}

// EventsForDay возвращает события на день
// @Summary События на день
// @Description Возвращает все события пользователя на указанный день
// @Tags events
// @Produce json
// @Param user_id query int true "ID пользователя"
// @Param date query string true "Дата в формате YYYY-MM-DD"
// @Success 200 {object} repository.SuccessResponse{result=repository.EventsResponse}
// @Failure 400 {object} repository.ErrorResponse
// @Router /events_for_day [get]
func (h *Handlers) EventsForDay(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	dateStr := r.URL.Query().Get("date")

	if userIDStr == "" || dateStr == "" {
		sendError(w, "user_id and date parameters are required", http.StatusBadRequest)
		return
	}

	userID, err := event.ValidateQueryParams(userIDStr, dateStr)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	date, err := event.ParseAndValidateDate(dateStr)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	events := h.serviceCalendar.GetEventsForDay(userID, date)
	sendResponse(w, repository.EventsResponse{Events: events}, http.StatusOK)
}

// EventsForWeek возвращает события на неделю
// @Summary События на неделю
// @Description Возвращает все события пользователя на указанную неделю
// @Tags events
// @Produce json
// @Param user_id query int true "ID пользователя"
// @Param date query string true "Дата в формате YYYY-MM-DD"
// @Success 200 {object} repository.SuccessResponse{result=repository.EventsResponse}
// @Failure 400 {object} repository.ErrorResponse
// @Router /events_for_week [get]
func (h *Handlers) EventsForWeek(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	dateStr := r.URL.Query().Get("date")

	if userIDStr == "" || dateStr == "" {
		sendError(w, "user_id and date parameters are required", http.StatusBadRequest)
		return
	}

	userID, err := event.ValidateQueryParams(userIDStr, dateStr)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	date, err := event.ParseAndValidateDate(dateStr)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	events := h.serviceCalendar.GetEventsForWeek(userID, date)
	sendResponse(w, repository.EventsResponse{Events: events}, http.StatusOK)
}

// EventsForMonth возвращает события на месяц
// @Summary События на месяц
// @Description Возвращает все события пользователя на указанный месяц
// @Tags events
// @Produce json
// @Param user_id query int true "ID пользователя"
// @Param date query string true "Дата в формате YYYY-MM-DD"
// @Success 200 {object} repository.SuccessResponse{result=repository.EventsResponse}
// @Failure 400 {object} repository.ErrorResponse
// @Router /events_for_month [get]
func (h *Handlers) EventsForMonth(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	dateStr := r.URL.Query().Get("date")

	if userIDStr == "" || dateStr == "" {
		sendError(w, "user_id and date parameters are required", http.StatusBadRequest)
		return
	}

	userID, err := event.ValidateQueryParams(userIDStr, dateStr)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	date, err := event.ParseAndValidateDate(dateStr)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	events := h.serviceCalendar.GetEventsForMonth(userID, date)
	sendResponse(w, repository.EventsResponse{Events: events}, http.StatusOK)
}

func sendResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)

	response := repository.SuccessResponse{Result: data}
	if err := encoder.Encode(response); err != nil {
		http.Error(w, `{"error": "failed to encode response"}`, http.StatusInternalServerError)
	}
}

func sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)

	response := repository.ErrorResponse{Error: message}
	if err := encoder.Encode(response); err != nil {
		http.Error(w, `{"error": "failed to encode error response"}`, http.StatusInternalServerError)
	}
}

// HealthCheck проверка здоровья сервера
// @Summary Проверка здоровья
// @Description Проверяет, что сервер работает
// @Tags utility
// @Produce json
// @Success 200 {object} repository.SuccessResponse{result=repository.SuccessResponse}
// @Router /health [get]
func (h *Handlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	sendResponse(w, map[string]string{"status": "ok", "timestamp": time.Now().Format(time.RFC3339)}, http.StatusOK)
}

func (h *Handlers) NotFound(w http.ResponseWriter, r *http.Request) {
	sendError(w, "route not found", http.StatusNotFound)
}
