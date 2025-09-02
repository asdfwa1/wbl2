package calendar

import (
	"calendar/internal/event/repository"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func testLogger() *slog.Logger {
	return slog.Default()
}

func TestCalendarService_Validation(t *testing.T) {
	repo := repository.NewEventRepository(testLogger())
	service := NewServiceCalendar(repo, testLogger())

	testDate := time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		title     string
		shouldErr bool
	}{
		{"empty title", "", true},
		{"whitespace title", "   ", true},
		{"tab characters", "\t\t\t", true},
		{"newline characters", "\n\n", true},
		{"mixed whitespace", " \t\n ", true},
		{"valid title", "Valid Title", false},
		{"valid with spaces", "  Valid with spaces  ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.CreateEvent(1, testDate, tt.title)
			if tt.shouldErr {
				assert.Error(t, err)
				assert.Equal(t, repository.ErrInvalidDataInput, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCalendarService_ErrorHandling(t *testing.T) {
	repo := repository.NewEventRepository(testLogger())
	service := NewServiceCalendar(repo, testLogger())

	testDate := time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)

	t.Run("update non-existent event", func(t *testing.T) {
		_, err := service.UpdateEvent(999, 1, testDate, "Title")
		assert.Error(t, err)
		assert.Equal(t, repository.ErrEventNotFound, err)
	})

	t.Run("delete non-existent event", func(t *testing.T) {
		err := service.DeleteEvent(999, 1)
		assert.Error(t, err)
		assert.Equal(t, repository.ErrEventNotFound, err)
	})

	t.Run("update event with wrong user", func(t *testing.T) {
		event, _ := service.CreateEvent(1, testDate, "Test Event")
		_, err := service.UpdateEvent(event.ID, 999, testDate, "New Title")
		assert.Error(t, err)
		assert.Equal(t, repository.ErrEventNotFound, err)
	})
}

func TestCalendarService_BusinessLogic(t *testing.T) {
	repo := repository.NewEventRepository(testLogger())
	service := NewServiceCalendar(repo, testLogger())

	testDate := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)

	t.Run("event lifecycle", func(t *testing.T) {
		event, err := service.CreateEvent(1, testDate, "Meeting")
		assert.NoError(t, err)
		assert.Equal(t, "Meeting", event.Title)

		events := service.GetEventsForDay(1, testDate)
		assert.Len(t, events, 1)
		assert.Equal(t, "Meeting", events[0].Title)

		updatedEvent, err := service.UpdateEvent(event.ID, event.UserID, testDate, "Updated Meeting")
		assert.NoError(t, err)
		assert.Equal(t, "Updated Meeting", updatedEvent.Title)

		events = service.GetEventsForDay(1, testDate)
		assert.Len(t, events, 1)
		assert.Equal(t, "Updated Meeting", events[0].Title)

		err = service.DeleteEvent(event.ID, event.UserID)
		assert.NoError(t, err)

		events = service.GetEventsForDay(1, testDate)
		assert.Empty(t, events)
	})
}

func TestCalendarService_ConcurrentValidation(t *testing.T) {
	repo := repository.NewEventRepository(testLogger())
	service := NewServiceCalendar(repo, testLogger())

	iterations := 20
	testDate := time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)

	t.Run("concurrent validation stress test", func(t *testing.T) {
		validDone := make(chan bool, iterations)
		invalidDone := make(chan bool, iterations)

		for i := 0; i < iterations; i++ {
			go func(userID int) {
				_, err := service.CreateEvent(userID, testDate, "Valid Event")
				assert.NoError(t, err)
				validDone <- true
			}(i % 10)
		}

		for i := 0; i < iterations; i++ {
			go func() {
				_, err := service.CreateEvent(1, testDate, "   ")
				assert.Error(t, err)
				invalidDone <- true
			}()
		}

		for i := 0; i < iterations*2; i++ {
			if i < iterations {
				<-validDone
			} else {
				<-invalidDone
			}
		}

		totalEvents := 0
		for userID := 0; userID < 10; userID++ {
			events := service.GetEventsForMonth(userID, testDate)
			totalEvents += len(events)
		}
		assert.Equal(t, iterations, totalEvents)
	})
}
