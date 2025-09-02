package repository

import (
	"errors"
	"log/slog"
	"sync"
	"time"
)

var (
	ErrEventNotFound    = errors.New("event not found")
	ErrInvalidDataInput = errors.New("invalid data input")
)

type EventRepository struct {
	mu     sync.RWMutex
	events []Event
	nextID int
	log    *slog.Logger
}

func NewEventRepository(logger *slog.Logger) *EventRepository {
	return &EventRepository{
		events: make([]Event, 20),
		nextID: 1,
		log:    logger,
	}
}

func (er *EventRepository) CreateEvent(userID int, date time.Time, title string) (Event, error) {
	er.mu.Lock()
	defer er.mu.Unlock()

	now := time.Now()
	event := Event{
		ID:        er.nextID,
		UserID:    userID,
		Date:      date,
		Title:     title,
		CreatedAt: now,
		UpdatedAt: now,
	}

	er.events = append(er.events, event)
	er.nextID++

	er.log.Info("Event created",
		"event_id", event.ID,
		"user_id", event.UserID,
		"date", event.Date.Format("2006-01-02"),
		"title", title,
	)

	return event, nil
}

func (er *EventRepository) GetEventsForDay(userID int, date time.Time) []Event {
	er.mu.RLock()
	defer er.mu.RUnlock()

	var result []Event
	for _, event := range er.events {
		if event.UserID == userID && sameDay(event.Date, date) {
			result = append(result, event)
		}
	}
	return result
}

func (er *EventRepository) GetEventsForWeek(userID int, date time.Time) []Event {
	er.mu.RLock()
	defer er.mu.RUnlock()

	var result []Event
	year, week := date.ISOWeek()
	for _, event := range er.events {
		if event.UserID == userID {
			eventYear, eventWeek := event.Date.ISOWeek()
			if eventYear == year && eventWeek == week {
				result = append(result, event)
			}
		}
	}
	return result
}

func (er *EventRepository) GetEventsForMonth(userID int, date time.Time) []Event {
	er.mu.RLock()
	defer er.mu.RUnlock()

	var result []Event
	year := date.Year()
	month := date.Month()
	for _, event := range er.events {
		if event.UserID == userID {
			eventYear := event.Date.Year()
			eventMonth := event.Date.Month()
			if eventYear == year && eventMonth == month {
				result = append(result, event)
			}
		}
	}
	return result
}

func (er *EventRepository) UpdateEvent(eventID, userID int, date time.Time, title string) (Event, error) {
	er.mu.Lock()
	defer er.mu.Unlock()

	for i, event := range er.events {
		if event.ID == eventID && event.UserID == userID {
			er.events[i].Date = date
			er.events[i].Title = title
			er.events[i].UpdatedAt = time.Now()

			er.log.Info("Event updated",
				"event_id", eventID,
				"user_id", userID,
				"new_title", title,
				"new_date", date.Format("2006-01-02"),
			)

			return er.events[i], nil
		}
	}

	return Event{}, ErrEventNotFound
}

func (er *EventRepository) DeleteEvent(eventID, userID int) error {
	er.mu.Lock()
	defer er.mu.Unlock()

	for i, event := range er.events {
		if event.ID == eventID && event.UserID == userID {
			er.events = append(er.events[:i], er.events[i+1:]...)

			er.log.Info("Event deleted",
				"event_id", eventID,
				"user_id", userID,
			)

			return nil
		}
	}

	return ErrEventNotFound
}

func sameDay(time1, time2 time.Time) bool {
	year1, month1, day1 := time1.Date()
	year2, month2, day2 := time2.Date()
	return year1 == year2 && month1 == month2 && day1 == day2
}
