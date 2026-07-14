package models

import "time"

// Meeting represents a meeting session
type Meeting struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Title       string    `json:"title" gorm:"not null"`
	Date        time.Time `json:"date" gorm:"not null"`
	Location    string    `json:"location"`
	OrganizerID uint      `json:"organizer_id"`
	Status      string    `json:"status" gorm:"default:'pending'"` // pending, processing, completed, failed
	AudioURL    string    `json:"audio_url"`
	Transcript  string    `json:"transcript" gorm:"type:text"`
	CleanedText string    `json:"cleaned_text" gorm:"type:text"`
	Summary     string    `json:"summary" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Participant represents a meeting participant
type Participant struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	MeetingID uint      `json:"meeting_id" gorm:"index;not null"`
	Name      string    `json:"name" gorm:"not null"`
	Email     string    `json:"email"`
	Role      string    `json:"role"` // host, speaker, attendee
	CreatedAt time.Time `json:"created_at"`
}

// DiscussionPoint represents a point discussed in the meeting
type DiscussionPoint struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	MeetingID uint   `json:"meeting_id" gorm:"index;not null"`
	Point     string `json:"point" gorm:"type:text;not null"`
	Speaker   string `json:"speaker"`
	Sequence  int    `json:"sequence"`
}

// Decision represents decisions made in the meeting
type Decision struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	MeetingID uint   `json:"meeting_id" gorm:"index;not null"`
	Decision  string `json:"decision" gorm:"type:text;not null"`
}

// ActionItem represents tasks assigned during the meeting
type ActionItem struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	MeetingID   uint      `json:"meeting_id" gorm:"index;not null"`
	Task        string    `json:"task" gorm:"type:text;not null"`
	Assignee    string    `json:"assignee"`
	Deadline    time.Time `json:"deadline"`
	Status      string    `json:"status" gorm:"default:'pending'"` // pending, in_progress, completed
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// MeetingMinute is the complete output combining all meeting data
type MeetingMinute struct {
	Meeting         Meeting          `json:"meeting"`
	Participants    []Participant    `json:"participants"`
	DiscussionPoints []DiscussionPoint `json:"discussion_points"`
	Decisions       []Decision       `json:"decisions"`
	ActionItems     []ActionItem     `json:"action_items"`
}