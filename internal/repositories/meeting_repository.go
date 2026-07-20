package repositories

import (
	"strings"
	"sync"
	"time"

	"github.com/yourusername/meeting-minutes-ai/internal/models"
	"gorm.io/gorm"
)

// MeetingRepository defines meeting data operations
type MeetingRepository interface {
	Create(meeting *models.Meeting) error
	FindByID(id uint) (*models.Meeting, error)
	FindAll(page, pageSize int, search string) ([]models.Meeting, int64, error)
	Update(meeting *models.Meeting) error
	Delete(id uint) error

	// Participants
	AddParticipant(participant *models.Participant) error
	GetParticipants(meetingID uint) ([]models.Participant, error)
	RemoveParticipant(id uint) error

	// Discussion Points
	AddDiscussionPoint(point *models.DiscussionPoint) error
	GetDiscussionPoints(meetingID uint) ([]models.DiscussionPoint, error)

	// Decisions
	AddDecision(decision *models.Decision) error
	GetDecisions(meetingID uint) ([]models.Decision, error)

	// Action Items
	AddActionItem(item *models.ActionItem) error
	GetActionItems(meetingID uint) ([]models.ActionItem, error)
	UpdateActionItem(item *models.ActionItem) error

	// Dashboard
	GetDashboardStats() (*models.Meeting, int64, int64, error)
}

// meetingRepository is a GORM-based implementation
type meetingRepository struct {
	db *gorm.DB
}

// NewMeetingRepository creates a new GORM meeting repository
func NewMeetingRepository(db *gorm.DB) MeetingRepository {
	return &meetingRepository{db: db}
}

func (r *meetingRepository) Create(meeting *models.Meeting) error {
	return r.db.Create(meeting).Error
}

func (r *meetingRepository) FindByID(id uint) (*models.Meeting, error) {
	var meeting models.Meeting
	err := r.db.First(&meeting, id).Error
	if err != nil {
		return nil, err
	}
	return &meeting, nil
}

func (r *meetingRepository) FindAll(page, pageSize int, search string) ([]models.Meeting, int64, error) {
	var meetings []models.Meeting
	var total int64

	query := r.db.Model(&models.Meeting{})

	if search != "" {
		query = query.Where("LOWER(title) LIKE LOWER(?) OR LOWER(location) LIKE LOWER(?) OR LOWER(transcript) LIKE LOWER(?) OR LOWER(summary) LIKE LOWER(?)",
			"%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&meetings).Error
	if err != nil {
		return nil, 0, err
	}

	return meetings, total, nil
}

func (r *meetingRepository) Update(meeting *models.Meeting) error {
	return r.db.Save(meeting).Error
}

func (r *meetingRepository) Delete(id uint) error {
	return r.db.Delete(&models.Meeting{}, id).Error
}

func (r *meetingRepository) AddParticipant(participant *models.Participant) error {
	return r.db.Create(participant).Error
}

func (r *meetingRepository) GetParticipants(meetingID uint) ([]models.Participant, error) {
	var participants []models.Participant
	err := r.db.Where("meeting_id = ?", meetingID).Find(&participants).Error
	return participants, err
}

func (r *meetingRepository) RemoveParticipant(id uint) error {
	return r.db.Delete(&models.Participant{}, id).Error
}

func (r *meetingRepository) AddDiscussionPoint(point *models.DiscussionPoint) error {
	return r.db.Create(point).Error
}

func (r *meetingRepository) GetDiscussionPoints(meetingID uint) ([]models.DiscussionPoint, error) {
	var points []models.DiscussionPoint
	err := r.db.Where("meeting_id = ?", meetingID).Order("sequence ASC").Find(&points).Error
	return points, err
}

func (r *meetingRepository) AddDecision(decision *models.Decision) error {
	return r.db.Create(decision).Error
}

func (r *meetingRepository) GetDecisions(meetingID uint) ([]models.Decision, error) {
	var decisions []models.Decision
	err := r.db.Where("meeting_id = ?", meetingID).Find(&decisions).Error
	return decisions, err
}

func (r *meetingRepository) AddActionItem(item *models.ActionItem) error {
	return r.db.Create(item).Error
}

func (r *meetingRepository) GetActionItems(meetingID uint) ([]models.ActionItem, error) {
	var items []models.ActionItem
	err := r.db.Where("meeting_id = ?", meetingID).Order("created_at DESC").Find(&items).Error
	return items, err
}

func (r *meetingRepository) UpdateActionItem(item *models.ActionItem) error {
	return r.db.Save(item).Error
}

func (r *meetingRepository) GetDashboardStats() (*models.Meeting, int64, int64, error) {
	var totalMeetings int64
	var completedActions int64
	var totalActions int64

	r.db.Model(&models.Meeting{}).Count(&totalMeetings)
	r.db.Model(&models.ActionItem{}).Count(&totalActions)
	r.db.Model(&models.ActionItem{}).Where("status = ?", "completed").Count(&completedActions)

	return nil, totalMeetings, completedActions, nil
}

// ==================== In-Memory Implementation ====================

// InMemoryMeetingRepository is an in-memory implementation for development
type InMemoryMeetingRepository struct {
	mu      sync.RWMutex
	meetings              map[uint]*models.Meeting
	participants          map[uint]*models.Participant
	discussionPoints      map[uint]*models.DiscussionPoint
	decisions             map[uint]*models.Decision
	actionItems           map[uint]*models.ActionItem
	meetingNextID         uint
	participantNextID     uint
	discussionPointNextID uint
	decisionNextID        uint
	actionItemNextID      uint
}

// NewInMemoryMeetingRepository creates a new in-memory meeting repository
func NewInMemoryMeetingRepository() MeetingRepository {
	return &InMemoryMeetingRepository{
		meetings:              make(map[uint]*models.Meeting),
		participants:          make(map[uint]*models.Participant),
		discussionPoints:      make(map[uint]*models.DiscussionPoint),
		decisions:             make(map[uint]*models.Decision),
		actionItems:           make(map[uint]*models.ActionItem),
		meetingNextID:         1,
		participantNextID:     1,
		discussionPointNextID: 1,
		decisionNextID:        1,
		actionItemNextID:      1,
	}
}

func (r *InMemoryMeetingRepository) Create(meeting *models.Meeting) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	meeting.ID = r.meetingNextID
	r.meetingNextID++
	now := time.Now()
	meeting.CreatedAt = now
	meeting.UpdatedAt = now
	r.meetings[meeting.ID] = meeting
	return nil
}

func (r *InMemoryMeetingRepository) FindByID(id uint) (*models.Meeting, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	meeting, exists := r.meetings[id]
	if !exists {
		return nil, gorm.ErrRecordNotFound
	}
	return meeting, nil
}

func (r *InMemoryMeetingRepository) FindAll(page, pageSize int, search string) ([]models.Meeting, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var meetings []models.Meeting
	for _, m := range r.meetings {
		if search != "" {
			searchLower := strings.ToLower(search)
			if !strings.Contains(strings.ToLower(m.Title), searchLower) &&
				!strings.Contains(strings.ToLower(m.Location), searchLower) &&
				!strings.Contains(strings.ToLower(m.Transcript), searchLower) &&
				!strings.Contains(strings.ToLower(m.Summary), searchLower) {
				continue
			}
		}
		meetings = append(meetings, *m)
	}

	total := int64(len(meetings))

	start := (page - 1) * pageSize
	if start >= len(meetings) {
		return []models.Meeting{}, total, nil
	}

	end := start + pageSize
	if end > len(meetings) {
		end = len(meetings)
	}

	return meetings[start:end], total, nil
}

func (r *InMemoryMeetingRepository) Update(meeting *models.Meeting) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.meetings[meeting.ID]
	if !exists {
		return gorm.ErrRecordNotFound
	}

	meeting.UpdatedAt = time.Now()
	r.meetings[meeting.ID] = meeting
	return nil
}

func (r *InMemoryMeetingRepository) Delete(id uint) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.meetings[id]
	if !exists {
		return gorm.ErrRecordNotFound
	}

	delete(r.meetings, id)
	return nil
}

func (r *InMemoryMeetingRepository) AddParticipant(participant *models.Participant) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	participant.ID = r.participantNextID
	r.participantNextID++
	now := time.Now()
	participant.CreatedAt = now
	r.participants[participant.ID] = participant
	return nil
}

func (r *InMemoryMeetingRepository) GetParticipants(meetingID uint) ([]models.Participant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var participants []models.Participant
	for _, p := range r.participants {
		if p.MeetingID == meetingID {
			participants = append(participants, *p)
		}
	}
	return participants, nil
}

func (r *InMemoryMeetingRepository) RemoveParticipant(id uint) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.participants[id]
	if !exists {
		return gorm.ErrRecordNotFound
	}
	delete(r.participants, id)
	return nil
}

func (r *InMemoryMeetingRepository) AddDiscussionPoint(point *models.DiscussionPoint) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	point.ID = r.discussionPointNextID
	r.discussionPointNextID++
	r.discussionPoints[point.ID] = point
	return nil
}

func (r *InMemoryMeetingRepository) GetDiscussionPoints(meetingID uint) ([]models.DiscussionPoint, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var points []models.DiscussionPoint
	for _, p := range r.discussionPoints {
		if p.MeetingID == meetingID {
			points = append(points, *p)
		}
	}
	return points, nil
}

func (r *InMemoryMeetingRepository) AddDecision(decision *models.Decision) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	decision.ID = r.decisionNextID
	r.decisionNextID++
	r.decisions[decision.ID] = decision
	return nil
}

func (r *InMemoryMeetingRepository) GetDecisions(meetingID uint) ([]models.Decision, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var decisions []models.Decision
	for _, d := range r.decisions {
		if d.MeetingID == meetingID {
			decisions = append(decisions, *d)
		}
	}
	return decisions, nil
}

func (r *InMemoryMeetingRepository) AddActionItem(item *models.ActionItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	item.ID = r.actionItemNextID
	r.actionItemNextID++
	now := time.Now()
	item.CreatedAt = now
	item.UpdatedAt = now
	r.actionItems[item.ID] = item
	return nil
}

func (r *InMemoryMeetingRepository) GetActionItems(meetingID uint) ([]models.ActionItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var items []models.ActionItem
	for _, item := range r.actionItems {
		if item.MeetingID == meetingID {
			items = append(items, *item)
		}
	}
	return items, nil
}

func (r *InMemoryMeetingRepository) UpdateActionItem(item *models.ActionItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.actionItems[item.ID]
	if !exists {
		return gorm.ErrRecordNotFound
	}

	item.UpdatedAt = time.Now()
	r.actionItems[item.ID] = item
	return nil
}

func (r *InMemoryMeetingRepository) GetDashboardStats() (*models.Meeting, int64, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	totalMeetings := int64(len(r.meetings))
	var completedActions, totalActions int64
	for _, item := range r.actionItems {
		totalActions++
		if item.Status == "completed" {
			completedActions++
		}
	}

	return nil, totalMeetings, completedActions, nil
}