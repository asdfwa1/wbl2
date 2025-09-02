package event

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidUserID  = errors.New("userID must be positive integer")
	ErrInvalidEventID = errors.New("eventID must be positive integer")
	ErrInvalidDate    = errors.New("date must be in YYYY-MM-DD format")
	ErrEmptyTitle     = errors.New("title cannot be empty")
	ErrTitleTooLong   = errors.New("title too long (max 255 characters)")
)

func ValidateCreateRequest(userID int, dateStr, title string) error {
	if userID <= 0 {
		return ErrInvalidUserID
	}

	if err := validateDate(dateStr); err != nil {
		return err
	}

	if err := validateTitle(title); err != nil {
		return err
	}

	return nil
}

func ValidateUpdateRequest(eventID, userID int, dateStr, title string) error {
	if eventID <= 0 {
		return ErrInvalidEventID
	}

	if userID <= 0 {
		return ErrInvalidUserID
	}

	if err := validateDate(dateStr); err != nil {
		return err
	}

	if err := validateTitle(title); err != nil {
		return err
	}

	return nil
}

func ValidateDeleteRequest(eventID, userID int) error {
	if eventID <= 0 {
		return ErrInvalidEventID
	}

	if userID <= 0 {
		return ErrInvalidUserID
	}

	return nil
}

func ValidateQueryParams(userIDStr, dateStr string) (int, error) {
	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID <= 0 {
		return 0, ErrInvalidUserID
	}

	if err := validateDate(dateStr); err != nil {
		return 0, err
	}

	return userID, nil
}

func validateDate(dateStr string) error {
	if dateStr == "" {
		return ErrInvalidDate
	}

	parts := strings.Split(dateStr, "-")
	if len(parts) != 3 {
		return ErrInvalidDate
	}

	for _, part := range parts {
		if _, err := strconv.Atoi(part); err != nil {
			return ErrInvalidDate
		}
	}

	_, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return ErrInvalidDate
	}

	return nil
}

func validateTitle(title string) error {
	if strings.TrimSpace(title) == "" {
		return ErrEmptyTitle
	}

	if len(title) > 255 {
		return ErrTitleTooLong
	}

	return nil
}

func ParseAndValidateDate(dateStr string) (time.Time, error) {
	if err := validateDate(dateStr); err != nil {
		return time.Time{}, err
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date: %w", err)
	}

	return date, nil
}
