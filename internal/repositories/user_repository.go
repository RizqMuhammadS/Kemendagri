package repositories

import (
	"sync"
	"time"

	"github.com/yourusername/meeting-minutes-ai/internal/models"
	"gorm.io/gorm"
)

// UserRepository defines user data operations
type UserRepository interface {
	Create(user *models.User) error
	FindByID(id uint) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	FindAll(page, pageSize int) ([]models.User, int64, error)
	Update(user *models.User) error
	Delete(id uint) error
}

// userRepository is a GORM-based implementation
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new GORM user repository
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) FindByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindAll(page, pageSize int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	r.db.Model(&models.User{}).Count(&total)
	offset := (page - 1) * pageSize
	err := r.db.Offset(offset).Limit(pageSize).Find(&users).Error
	if err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *userRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) Delete(id uint) error {
	return r.db.Delete(&models.User{}, id).Error
}

// ==================== In-Memory Implementation ====================

// InMemoryUserRepository is an in-memory implementation for development
type InMemoryUserRepository struct {
	mu     sync.RWMutex
	users  map[uint]*models.User
	nextID uint
}

// NewInMemoryUserRepository creates a new in-memory user repository
func NewInMemoryUserRepository() UserRepository {
	return &InMemoryUserRepository{
		users:  make(map[uint]*models.User),
		nextID: 1,
	}
}

func (r *InMemoryUserRepository) Create(user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user.ID = r.nextID
	r.nextID++
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now
	r.users[user.ID] = user
	return nil
}

func (r *InMemoryUserRepository) FindByID(id uint) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, gorm.ErrRecordNotFound
	}
	return user, nil
}

func (r *InMemoryUserRepository) FindByEmail(email string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *InMemoryUserRepository) FindAll(page, pageSize int) ([]models.User, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var users []models.User
	for _, user := range r.users {
		users = append(users, *user)
	}

	total := int64(len(users))

	// Simple pagination
	start := (page - 1) * pageSize
	if start >= len(users) {
		return []models.User{}, total, nil
	}

	end := start + pageSize
	if end > len(users) {
		end = len(users)
	}

	return users[start:end], total, nil
}

func (r *InMemoryUserRepository) Update(user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.users[user.ID]
	if !exists {
		return gorm.ErrRecordNotFound
	}

	user.UpdatedAt = time.Now()
	r.users[user.ID] = user
	return nil
}

func (r *InMemoryUserRepository) Delete(id uint) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.users[id]
	if !exists {
		return gorm.ErrRecordNotFound
	}

	delete(r.users, id)
	return nil
}