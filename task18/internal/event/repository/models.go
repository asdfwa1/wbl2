package repository

import (
	"time"
)

type Event struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Title     string    `json:"title"`
	Date      time.Time `json:"date"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateEventRequest struct {
	UserID int    `json:"user_id" example:"1" binding:"required"`
	Date   string `json:"date" example:"YYYY-MM-DD" binding:"required" format:"date"`
	Title  string `json:"title" example:"example string" binding:"required"`
}

type UpdateEventRequest struct {
	EventID int    `json:"event_id" example:"1" binding:"required"`
	UserID  int    `json:"user_id" example:"1" binding:"required"`
	Date    string `json:"date" example:"YYYY-MM-DD" binding:"required" format:"date"`
	Title   string `json:"title" example:"example string" binding:"required"`
}

type DeleteEventRequest struct {
	EventID int `json:"event_id" example:"1" binding:"required"`
	UserID  int `json:"user_id" example:"1" binding:"required"`
}

type EventsResponse struct {
	Events []Event `json:"events"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Result interface{} `json:"result"`
}
