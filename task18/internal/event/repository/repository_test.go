package repository

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"log/slog"
)

func testLogger() *slog.Logger {
	return slog.Default()
}

func TestEventRepository_CreateEvent(t *testing.T) {
	repo := NewEventRepository(testLogger())

	t.Run("successful event creation", func(t *testing.T) {
		date := time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)
		event, err := repo.CreateEvent(1, date, "New Party")

		assert.NoError(t, err)
		assert.Equal(t, 1, event.ID)
		assert.Equal(t, 1, event.UserID)
		assert.Equal(t, "New Party", event.Title)
		assert.Equal(t, date, event.Date)
		assert.False(t, event.CreatedAt.IsZero())
		assert.False(t, event.UpdatedAt.IsZero())
	})

	t.Run("events have incremental IDs", func(t *testing.T) {
		date := time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)

		event1, err1 := repo.CreateEvent(1, date, "Event 1")
		event2, err2 := repo.CreateEvent(2, date, "Event 2")

		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.Equal(t, 2, event1.ID)
		assert.Equal(t, 3, event2.ID)
		assert.Equal(t, 1, event1.UserID)
		assert.Equal(t, 2, event2.UserID)
	})
}

func TestEventRepository_UpdateEvent(t *testing.T) {
	repo := NewEventRepository(testLogger())

	t.Run("successful event update", func(t *testing.T) {
		oldDate := time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)
		event, err := repo.CreateEvent(1, oldDate, "Old Title")
		assert.NoError(t, err)

		time.Sleep(1 * time.Millisecond)

		newDate := time.Date(2025, 9, 2, 0, 0, 0, 0, time.UTC)
		updatedEvent, err := repo.UpdateEvent(event.ID, event.UserID, newDate, "Updated Title")

		assert.NoError(t, err)
		assert.Equal(t, event.ID, updatedEvent.ID)
		assert.Equal(t, event.UserID, updatedEvent.UserID)
		assert.Equal(t, "Updated Title", updatedEvent.Title)
		assert.Equal(t, newDate, updatedEvent.Date)
		assert.True(t, updatedEvent.UpdatedAt.After(event.UpdatedAt))
	})

	t.Run("update non-existent event", func(t *testing.T) {
		newDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		_, err := repo.UpdateEvent(999, 1, newDate, "Title")

		assert.Error(t, err)
		assert.Equal(t, ErrEventNotFound, err)
	})

	t.Run("update event with wrong user id", func(t *testing.T) {
		date := time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)
		event, err := repo.CreateEvent(1, date, "Test Event")
		assert.NoError(t, err)

		_, err = repo.UpdateEvent(event.ID, 999, date, "New Title")

		assert.Error(t, err)
		assert.Equal(t, ErrEventNotFound, err)
	})
}

func TestEventRepository_DeleteEvent(t *testing.T) {
	repo := NewEventRepository(testLogger())

	t.Run("successful event deletion", func(t *testing.T) {
		date := time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)
		event, err := repo.CreateEvent(1, date, "Event to delete")
		assert.NoError(t, err)

		err = repo.DeleteEvent(event.ID, event.UserID)
		assert.NoError(t, err)

		events := repo.GetEventsForDay(1, date)
		assert.Empty(t, events)
	})

	t.Run("delete non-existent event", func(t *testing.T) {
		err := repo.DeleteEvent(999, 1)
		assert.Error(t, err)
		assert.Equal(t, ErrEventNotFound, err)
	})

	t.Run("delete event with wrong user id", func(t *testing.T) {
		date := time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)
		event, err := repo.CreateEvent(1, date, "Test Event")
		assert.NoError(t, err)

		err = repo.DeleteEvent(event.ID, 999)
		assert.Error(t, err)
		assert.Equal(t, ErrEventNotFound, err)
	})
}

func TestEventRepository_GetEventsForPeriod(t *testing.T) {
	repo := NewEventRepository(testLogger())

	testDate := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	sameWeekDate := time.Date(2024, 12, 27, 0, 0, 0, 0, time.UTC)
	sameMonthDate := time.Date(2024, 12, 12, 0, 0, 0, 0, time.UTC)
	nextMonthDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	repo.CreateEvent(1, testDate, "User 1 Event 1")
	repo.CreateEvent(1, testDate, "User 1 Event 2")
	repo.CreateEvent(1, sameWeekDate, "User 1 Same Week")
	repo.CreateEvent(1, sameMonthDate, "User 1 Same Month")
	repo.CreateEvent(1, nextMonthDate, "User 1 Next Month")
	repo.CreateEvent(2, testDate, "User 2 Event")

	t.Run("get events for day", func(t *testing.T) {
		events := repo.GetEventsForDay(1, testDate)

		assert.Len(t, events, 2)
		for _, event := range events {
			assert.Equal(t, 1, event.UserID)
			assert.True(t, sameDay(event.Date, testDate))
		}
	})

	t.Run("get events for week", func(t *testing.T) {
		events := repo.GetEventsForWeek(1, testDate)
		assert.Len(t, events, 3)
	})

	t.Run("get events for month", func(t *testing.T) {
		events := repo.GetEventsForMonth(1, testDate)
		assert.Len(t, events, 4)
	})

	t.Run("get events for different user", func(t *testing.T) {
		events := repo.GetEventsForDay(2, testDate)
		assert.Len(t, events, 1)
		assert.Equal(t, 2, events[0].UserID)
	})

	t.Run("get events for non-existent user", func(t *testing.T) {
		events := repo.GetEventsForDay(999, testDate)
		assert.Empty(t, events)
	})
}

func TestEventRepository_ConcurrentAccess(t *testing.T) {
	repo := NewEventRepository(testLogger())
	iterations := 100
	date := time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)

	t.Run("concurrent create events", func(t *testing.T) {
		done := make(chan bool, iterations)
		for i := 0; i < iterations; i++ {
			go func(index int) {
				userID := (index % 10) + 1
				_, err := repo.CreateEvent(userID, date, "Concurrent Event")
				assert.NoError(t, err)
				done <- true
			}(i)
		}

		for i := 0; i < iterations; i++ {
			<-done
		}

		totalEvents := 0
		for userID := 1; userID <= 10; userID++ {
			events := repo.GetEventsForMonth(userID, date)
			totalEvents += len(events)
		}

		assert.Equal(t, iterations, totalEvents)
	})
}
