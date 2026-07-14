package dto

import (
	"time"
)

// APIResponse is the standard API response wrapper
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// PaginationResponse holds pagination metadata
type PaginationResponse struct {
	Page      int `json:"page"`
	PageSize  int `json:"page_size"`
	Total     int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// MeetingResponse represents a meeting in API responses
type MeetingResponse struct {
	ID            uint      `json:"id"`
	Title         string    `json:"title"`
	Date          string    `json:"date"`
	Location      string    `json:"location"`
	Status        string    `json:"status"`
	ParticipantCount int    `json:"participant_count"`
	CreatedAt     time.Time `json:"created_at"`
}

// MeetingDetailResponse is the full meeting detail
type MeetingDetailResponse struct {
	ID              uint                 `json:"id"`
	Title           string               `json:"title"`
	Date            string               `json:"date"`
	Location        string               `json:"location"`
	Status          string               `json:"status"`
	OrganizerID     uint                 `json:"organizer_id"`
	AudioURL        string               `json:"audio_url"`
	Transcript      string               `json:"transcript"`
	CleanedText     string               `json:"cleaned_text"`
	Summary         string               `json:"summary"`
	Participants    []ParticipantResponse `json:"participants"`
	DiscussionPoints []DiscussionPointResponse `json:"discussion_points"`
	Decisions       []DecisionResponse    `json:"decisions"`
	ActionItems     []ActionItemResponse  `json:"action_items"`
	CreatedAt       time.Time             `json:"created_at"`
	UpdatedAt       time.Time             `json:"updated_at"`
}

// ParticipantResponse represents a participant in responses
type ParticipantResponse struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// DiscussionPointResponse represents a discussion point in responses
type DiscussionPointResponse struct {
	ID       uint   `json:"id"`
	Point    string `json:"point"`
	Speaker  string `json:"speaker"`
	Sequence int    `json:"sequence"`
}

// DecisionResponse represents a decision in responses
type DecisionResponse struct {
	ID       uint   `json:"id"`
	Decision string `json:"decision"`
}

// ActionItemResponse represents an action item in responses
type ActionItemResponse struct {
	ID        uint      `json:"id"`
	Task      string    `json:"task"`
	Assignee  string    `json:"assignee"`
	Deadline  string    `json:"deadline"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	Token string `json:"token"`
	User  UserResponse `json:"user"`
}

// UserResponse represents user data in responses
type UserResponse struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// DashboardResponse contains summary statistics for dashboard
type DashboardResponse struct {
	TotalMeetings      int `json:"total_meetings"`
	CompletedMeetings  int `json:"completed_meetings"`
	PendingMeetings    int `json:"pending_meetings"`
	TotalActionItems   int `json:"total_action_items"`
	CompletedActions   int `json:"completed_actions"`
	PendingActions     int `json:"pending_actions"`
	TotalParticipants  int `json:"total_participants"`
}