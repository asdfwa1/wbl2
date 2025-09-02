package calendar

import (
	"calendar/internal/event/repository"
	"log/slog"
	"strings"
	"time"
)

type ServiceCalendar struct {
	repo *repository.EventRepository
	log  *slog.Logger
}

func NewServiceCalendar(repo *repository.EventRepository, logger *slog.Logger) *ServiceCalendar {
	return &ServiceCalendar{
		repo: repo,
		log:  logger,
	}
}

func (sc *ServiceCalendar) CreateEvent(userID int, date time.Time, title string) (repository.Event, error) {
	if strings.TrimSpace(title) == "" {
		return repository.Event{}, repository.ErrInvalidDataInput
	}
	return sc.repo.CreateEvent(userID, date, title)
}

func (sc *ServiceCalendar) UpdateEvent(eventID, userID int, date time.Time, title string) (repository.Event, error) {
	if strings.TrimSpace(title) == "" {
		return repository.Event{}, repository.ErrInvalidDataInput
	}
	return sc.repo.UpdateEvent(eventID, userID, date, title)
}

func (sc *ServiceCalendar) DeleteEvent(eventID, userID int) error {
	return sc.repo.DeleteEvent(eventID, userID)
}

func (sc *ServiceCalendar) GetEventsForDay(userID int, date time.Time) []repository.Event {
	return sc.repo.GetEventsForDay(userID, date)
}

func (sc *ServiceCalendar) GetEventsForWeek(userID int, date time.Time) []repository.Event {
	return sc.repo.GetEventsForWeek(userID, date)
}

func (sc *ServiceCalendar) GetEventsForMonth(userID int, date time.Time) []repository.Event {
	return sc.repo.GetEventsForMonth(userID, date)
}
