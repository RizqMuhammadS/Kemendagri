package dto

import "mime/multipart"

// CreateMeetingRequest represents request to create a new meeting
type CreateMeetingRequest struct {
	Title    string   `json:"title" validate:"required"`
	Date     string   `json:"date" validate:"required"`
	Location string   `json:"location"`
	Participants []ParticipantRequest `json:"participants"`
}

// UpdateMeetingRequest represents request to update a meeting's basic info
type UpdateMeetingRequest struct {
	Title    string `json:"title"`
	Date     string `json:"date"`
	Location string `json:"location"`
	Status   string `json:"status"`
}

// ParticipantRequest represents a participant in the request
type ParticipantRequest struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// UploadAudioRequest represents uploading audio file
type UploadAudioRequest struct {
	MeetingID uint                  `form:"meeting_id" validate:"required"`
	Audio     *multipart.FileHeader `form:"audio" validate:"required"`
}

// ProcessTranscriptRequest represents request to process transcript
type ProcessTranscriptRequest struct {
	MeetingID  uint   `json:"meeting_id" validate:"required"`
	Transcript string `json:"transcript" validate:"required"`
}

// LoginRequest represents login credentials
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// RegisterRequest represents registration request
type RegisterRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Role     string `json:"role"`
}

// ExportRequest represents request to export meeting minutes
type ExportRequest struct {
	MeetingID uint   `json:"meeting_id" validate:"required"`
	Format    string `json:"format" validate:"required"` // pdf, word
}

// SendEmailRequest represents request to send email
type SendEmailRequest struct {
	MeetingID uint     `json:"meeting_id" validate:"required"`
	Recipients []string `json:"recipients" validate:"required"`
	Format    string   `json:"format"` // pdf, word (attachment format)
}