package services

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/yourusername/meeting-minutes-ai/internal/config"
	"github.com/yourusername/meeting-minutes-ai/internal/dto"
	"github.com/yourusername/meeting-minutes-ai/internal/models"
	"github.com/yourusername/meeting-minutes-ai/internal/repositories"
)

// MeetingService orchestrates the meeting minutes workflow
type MeetingService struct {
	meetingRepo  repositories.MeetingRepository
	userRepo     repositories.UserRepository
	sttService   *STTService
	cleaner      *TextCleanerService
	llmService   *LLMService
	exportSvc    *ExportService
	emailSvc     *EmailService
	cfg          *config.Config
}

// NewMeetingService creates a new meeting service
func NewMeetingService(
	meetingRepo repositories.MeetingRepository,
	userRepo repositories.UserRepository,
	sttService *STTService,
	cleaner *TextCleanerService,
	llmService *LLMService,
	exportSvc *ExportService,
	emailSvc *EmailService,
	cfg *config.Config,
) *MeetingService {
	return &MeetingService{
		meetingRepo: meetingRepo,
		userRepo:    userRepo,
		sttService:  sttService,
		cleaner:     cleaner,
		llmService:  llmService,
		exportSvc:   exportSvc,
		emailSvc:    emailSvc,
		cfg:         cfg,
	}
}

// CreateMeeting creates a new meeting with participants
func (s *MeetingService) CreateMeeting(req *dto.CreateMeetingRequest, organizerID uint) (*models.Meeting, error) {
	meetingDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format, use YYYY-MM-DD: %w", err)
	}

	meeting := &models.Meeting{
		Title:       req.Title,
		Date:        meetingDate,
		Location:    req.Location,
		OrganizerID: organizerID,
		Status:      "pending",
	}

	if err := s.meetingRepo.Create(meeting); err != nil {
		return nil, fmt.Errorf("failed to create meeting: %w", err)
	}

	// Add participants
	for _, p := range req.Participants {
		participant := &models.Participant{
			MeetingID: meeting.ID,
			Name:      p.Name,
			Email:     p.Email,
			Role:      p.Role,
		}
		if err := s.meetingRepo.AddParticipant(participant); err != nil {
			return nil, fmt.Errorf("failed to add participant: %w", err)
		}
	}

	return meeting, nil
}

// UploadAudio handles audio file upload and triggers transcription via STT (Whisper API)
func (s *MeetingService) UploadAudio(meetingID uint, file *multipart.FileHeader) error {
	// Find meeting
	meeting, err := s.meetingRepo.FindByID(meetingID)
	if err != nil {
		return fmt.Errorf("meeting not found: %w", err)
	}

	// Save file to disk
	filename := fmt.Sprintf("%d_%s", meetingID, file.Filename)
	savePath := filepath.Join(s.cfg.UploadDir, filename)

	// Ensure upload directory exists
	if err := os.MkdirAll(s.cfg.UploadDir, 0755); err != nil {
		return fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(savePath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy file contents
	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	// Update meeting with audio URL
	meeting.AudioURL = savePath
	meeting.Status = "processing"
	_ = s.meetingRepo.Update(meeting)

	// Transcribe audio using STT service (OpenAI Whisper API)
	transcript, err := s.sttService.Transcribe(savePath)
	if err != nil {
		meeting.Status = "failed"
		_ = s.meetingRepo.Update(meeting)
		return fmt.Errorf("transkripsi audio gagal: %w", err)
	}

	// Auto-generate notulensi dari hasil transkrip menggunakan AI
	// Proses: transcript → cleaner → LLM summary → structured result
	_, err = s.ProcessTranscript(meetingID, transcript)
	if err != nil {
		// Jika AI gagal, transcript tetap tersimpan agar user bisa coba lagi
		meeting.Transcript = transcript
		meeting.Status = "transcribed"
		_ = s.meetingRepo.Update(meeting)
		return fmt.Errorf("transkripsi berhasil, tetapi generate notulensi AI gagal: %w", err)
	}

	return nil
}

// ProcessTranscript processes the transcript through cleaning and AI summarization
func (s *MeetingService) ProcessTranscript(meetingID uint, transcript string) (*models.Meeting, error) {
	meeting, err := s.meetingRepo.FindByID(meetingID)
	if err != nil {
		return nil, fmt.Errorf("meeting not found: %w", err)
	}

	// Save original transcript
	meeting.Transcript = transcript
	meeting.Status = "processing"
	_ = s.meetingRepo.Update(meeting)

	// Step 1: Clean the text
	cleanedText := s.cleaner.CleanIndonesian(transcript)
	meeting.CleanedText = cleanedText
	_ = s.meetingRepo.Update(meeting)

	// Step 2: Get participants for context
	participants, err := s.meetingRepo.GetParticipants(meetingID)
	if err != nil {
		participants = []models.Participant{}
	}

	// Step 3: Generate AI summary
	result, err := s.llmService.GenerateMinutes(meeting.Title, participants, cleanedText)
	if err != nil {
		meeting.Status = "failed"
		_ = s.meetingRepo.Update(meeting)
		return nil, fmt.Errorf("AI summarization failed: %w", err)
	}

	// Step 4: Save structured results
	meeting.Summary = result.Summary
	meeting.Status = "completed"

	// Save discussion points
	for i, point := range result.DiscussionPoints {
		dp := &models.DiscussionPoint{
			MeetingID: meetingID,
			Point:     point,
			Sequence:  i + 1,
		}
		_ = s.meetingRepo.AddDiscussionPoint(dp)
	}

	// Save decisions
	for _, decision := range result.Decisions {
		d := &models.Decision{
			MeetingID: meetingID,
			Decision:  decision,
		}
		_ = s.meetingRepo.AddDecision(d)
	}

	// Save action items
	for _, item := range result.ActionItems {
		var deadline time.Time
		if item.Deadline != "" {
			deadline, _ = time.Parse("2006-01-02", item.Deadline)
		}

		ai := &models.ActionItem{
			MeetingID: meetingID,
			Task:      item.Task,
			Assignee:  item.Assignee,
			Deadline:  deadline,
			Status:    "pending",
		}
		_ = s.meetingRepo.AddActionItem(ai)
	}

	if err := s.meetingRepo.Update(meeting); err != nil {
		return nil, fmt.Errorf("failed to update meeting: %w", err)
	}

	return meeting, nil
}

// GetMeetingDetail returns complete meeting details with all related data
func (s *MeetingService) GetMeetingDetail(meetingID uint) (*dto.MeetingDetailResponse, error) {
	meeting, err := s.meetingRepo.FindByID(meetingID)
	if err != nil {
		return nil, fmt.Errorf("meeting not found: %w", err)
	}

	participants, _ := s.meetingRepo.GetParticipants(meetingID)
	discussionPoints, _ := s.meetingRepo.GetDiscussionPoints(meetingID)
	decisions, _ := s.meetingRepo.GetDecisions(meetingID)
	actionItems, _ := s.meetingRepo.GetActionItems(meetingID)

	return &dto.MeetingDetailResponse{
		ID:          meeting.ID,
		Title:       meeting.Title,
		Date:        meeting.Date.Format("2006-01-02"),
		Location:    meeting.Location,
		Status:      meeting.Status,
		OrganizerID: meeting.OrganizerID,
		AudioURL:    meeting.AudioURL,
		Transcript:  meeting.Transcript,
		CleanedText: meeting.CleanedText,
		Summary:     meeting.Summary,
		Participants: toParticipantResponses(participants),
		DiscussionPoints: toDiscussionPointResponses(discussionPoints),
		Decisions:    toDecisionResponses(decisions),
		ActionItems:  toActionItemResponses(actionItems),
		CreatedAt:    meeting.CreatedAt,
		UpdatedAt:    meeting.UpdatedAt,
	}, nil
}

// ListMeetings returns paginated meetings with optional search
func (s *MeetingService) ListMeetings(page, pageSize int, search string) ([]dto.MeetingResponse, int64, error) {
	meetings, total, err := s.meetingRepo.FindAll(page, pageSize, search)
	if err != nil {
		return nil, 0, err
	}

	var responses []dto.MeetingResponse
	for _, m := range meetings {
		participants, _ := s.meetingRepo.GetParticipants(m.ID)
		responses = append(responses, dto.MeetingResponse{
			ID:               m.ID,
			Title:            m.Title,
			Date:             m.Date.Format("2006-01-02"),
			Location:         m.Location,
			Status:           m.Status,
			ParticipantCount: len(participants),
			CreatedAt:        m.CreatedAt,
		})
	}

	return responses, total, nil
}

// ExportMeeting exports meeting minutes in specified format
func (s *MeetingService) ExportMeeting(meetingID uint, format string) (string, error) {
	detail, err := s.GetMeetingDetail(meetingID)
	if err != nil {
		return "", err
	}

	return s.exportSvc.Export(detail, format)
}

// SendMeetingEmail sends meeting minutes via email
func (s *MeetingService) SendMeetingEmail(meetingID uint, recipients []string, format string) error {
	detail, err := s.GetMeetingDetail(meetingID)
	if err != nil {
		return err
	}

	attachmentPath, err := s.exportSvc.Export(detail, format)
	if err != nil {
		return fmt.Errorf("failed to export for email: %w", err)
	}

	// Check if SMTP is actually configured before trying to send
	if s.cfg.SMTPUser == "" || s.cfg.SMTPPass == "" {
		// SMTP not configured - simulate success and return a clear message
		// The file is still exported as attachment for manual download
		return fmt.Errorf("Email tidak dapat dikirim: SMTP belum dikonfigurasi. Silakan atur SMTP_USER dan SMTP_PASS di file .env. File export tersimpan di: %s", attachmentPath)
	}

	return s.emailSvc.SendMinutes(detail.Title, recipients, attachmentPath)
}

// UpdateMeeting updates a meeting's basic info (title, date, location, status)
func (s *MeetingService) UpdateMeeting(meetingID uint, req *dto.UpdateMeetingRequest) (*models.Meeting, error) {
	meeting, err := s.meetingRepo.FindByID(meetingID)
	if err != nil {
		return nil, fmt.Errorf("meeting not found: %w", err)
	}

	if req.Title != "" {
		meeting.Title = req.Title
	}
	if req.Date != "" {
		parsedDate, err := time.Parse("2006-01-02", req.Date)
		if err != nil {
			return nil, fmt.Errorf("invalid date format, use YYYY-MM-DD: %w", err)
		}
		meeting.Date = parsedDate
	}
	if req.Location != "" {
		meeting.Location = req.Location
	}
	if req.Status != "" {
		// Validate status
		validStatuses := map[string]bool{"pending": true, "processing": true, "completed": true, "failed": true, "transcribed": true}
		if !validStatuses[req.Status] {
			return nil, fmt.Errorf("invalid status: %s (valid: pending, processing, completed, failed, transcribed)", req.Status)
		}
		meeting.Status = req.Status
	}

	if err := s.meetingRepo.Update(meeting); err != nil {
		return nil, fmt.Errorf("failed to update meeting: %w", err)
	}

	return meeting, nil
}

// DeleteMeeting deletes a meeting by ID
func (s *MeetingService) DeleteMeeting(meetingID uint) error {
	// Check if meeting exists
	_, err := s.meetingRepo.FindByID(meetingID)
	if err != nil {
		return fmt.Errorf("meeting not found: %w", err)
	}

	return s.meetingRepo.Delete(meetingID)
}

// GetDashboardStats returns dashboard statistics
func (s *MeetingService) GetDashboardStats() (*dto.DashboardResponse, error) {
	// Get counts
	allMeetings, totalMeetings, _ := s.meetingRepo.FindAll(1, 10000, "")

	var completedCount, pendingCount int64
	for _, m := range allMeetings {
		if m.Status == "completed" {
			completedCount++
		} else if m.Status == "pending" {
			pendingCount++
		}
	}

	// Get total participants across all meetings
	var totalParticipants int64
	// Count distinct participants (simplified)
	for _, m := range allMeetings {
		participants, _ := s.meetingRepo.GetParticipants(m.ID)
		totalParticipants += int64(len(participants))
	}

	// Get action items stats
	var completedActions, pendingActions int64
	for _, m := range allMeetings {
		items, _ := s.meetingRepo.GetActionItems(m.ID)
		for _, item := range items {
			if item.Status == "completed" {
				completedActions++
			} else {
				pendingActions++
			}
		}
	}

	return &dto.DashboardResponse{
		TotalMeetings:      int(totalMeetings),
		CompletedMeetings:  int(completedCount),
		PendingMeetings:    int(pendingCount),
		TotalActionItems:   int(completedActions + pendingActions),
		CompletedActions:   int(completedActions),
		PendingActions:     int(pendingActions),
		TotalParticipants:  int(totalParticipants),
	}, nil
}

// Helper conversion functions
func toParticipantResponses(participants []models.Participant) []dto.ParticipantResponse {
	var result []dto.ParticipantResponse
	for _, p := range participants {
		result = append(result, dto.ParticipantResponse{
			ID:    p.ID,
			Name:  p.Name,
			Email: p.Email,
			Role:  p.Role,
		})
	}
	return result
}

func toDiscussionPointResponses(points []models.DiscussionPoint) []dto.DiscussionPointResponse {
	var result []dto.DiscussionPointResponse
	for _, p := range points {
		result = append(result, dto.DiscussionPointResponse{
			ID:       p.ID,
			Point:    p.Point,
			Speaker:  p.Speaker,
			Sequence: p.Sequence,
		})
	}
	return result
}

func toDecisionResponses(decisions []models.Decision) []dto.DecisionResponse {
	var result []dto.DecisionResponse
	for _, d := range decisions {
		result = append(result, dto.DecisionResponse{
			ID:       d.ID,
			Decision: d.Decision,
		})
	}
	return result
}

func toActionItemResponses(items []models.ActionItem) []dto.ActionItemResponse {
	var result []dto.ActionItemResponse
	for _, item := range items {
		deadlineStr := ""
		if !item.Deadline.IsZero() {
			deadlineStr = item.Deadline.Format("2006-01-02")
		}
		result = append(result, dto.ActionItemResponse{
			ID:        item.ID,
			Task:      item.Task,
			Assignee:  item.Assignee,
			Deadline:  deadlineStr,
			Status:    item.Status,
			CreatedAt: item.CreatedAt,
		})
	}
	return result
}