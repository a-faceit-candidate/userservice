package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/a-faceit-candidate/userservice/internal/model"
	"github.com/a-faceit-candidate/userservice/internal/persistence"
	"github.com/google/uuid"
)

// Service describes the functionality of the business logic in this service
type Service interface {
	// Create assumes that ID, CreatedAt, and UpdatedAt fields are empty.
	// Create will modify the user provided.
	Create(context.Context, *model.User) (*model.User, error)
	// Update will modify the user provided updating the UpdatedAt timestamp, and the ID will be set to the one provided
	Update(ctx context.Context, id string, user *model.User) (*model.User, error)
	Get(context.Context, string) (*model.User, error)
	Delete(context.Context, string) error
	ListAll(context.Context) ([]*model.User, error)
	ListCountry(ctx context.Context, countryCode string) ([]*model.User, error)
}

//go:generate mockery -output servicemock -outpkg servicemock -case underscore -name Service

var ErrNotFound = errors.New("not found")
var ErrConflict = errors.New("conflict updating")

// ErrInvalidParams won't be returned itself, but it will be wrapped by another error instead
// this is a shorthand to implemeting an own error type
var ErrInvalidParams = errors.New("invalid params")

// New provices an implementation of Service
func New(repo persistence.Repository) *ServiceImpl {
	return &ServiceImpl{
		repo: repo,
	}
}

// ServiceImpl is the default, and hopefully unique implementation of the service.
type ServiceImpl struct {
	repo persistence.Repository
}

// Create will fill the ID, CreatedAt and Updated at fields of the user before creating it
// it will also validate the fields, failing with an error wrapping ErrInvalidParams if some of them is invalid
// Notice that the timestamps have to be truncated to the microsecond, as that's the best precision mysql suports on the
// DATETIME field. Otherwise we'd be returning to the caller timestamp with more precision than stored, and they wouldn't
// be able to update the model using those timestamps later.
func (s *ServiceImpl) Create(ctx context.Context, user *model.User) (*model.User, error) {
	if err := s.validateUserForCreate(user); err != nil {
		return nil, err
	}
	s.replacePasswordByHash(user)

	user.ID = uuidv1()
	user.CreatedAt = timeNow().Truncate(time.Microsecond)
	user.UpdatedAt = timeNow().Truncate(time.Microsecond)

	if err := s.repo.Create(ctx, user); err != nil {
		if err == persistence.ErrConflict {
			// repository did it okay, but we failed at uniqueness
			return nil, fmt.Errorf("internal error: we've generated a duplicated uuid")
		}
		return nil, err
	}

	return user, nil
}

func (s *ServiceImpl) Update(ctx context.Context, id string, user *model.User) (*model.User, error) {
	if err := s.validateUserForUpdate(user); err != nil {
		return nil, err
	}
	if user.Password != "" {
		s.replacePasswordByHash(user)
	}

	providedUpdatedAt := user.UpdatedAt
	user.ID = id
	user.UpdatedAt = timeNow().Truncate(time.Microsecond)

	if err := s.repo.Update(ctx, user, providedUpdatedAt); err != nil {
		if err == persistence.ErrConflict {
			return nil, ErrConflict
		} else if err == persistence.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return user, nil
}

func (s *ServiceImpl) Get(ctx context.Context, id string) (*model.User, error) {
	user, err := s.repo.Get(ctx, id)
	if err != nil {
		if err == persistence.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return user, nil
}

func (s *ServiceImpl) Delete(ctx context.Context, id string) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		if err == persistence.ErrNotFound {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (s *ServiceImpl) ListAll(ctx context.Context) ([]*model.User, error) {
	return s.removePasswords(s.repo.ListAll(ctx))
}

func (s *ServiceImpl) ListCountry(ctx context.Context, countryCode string) ([]*model.User, error) {
	return s.removePasswords(s.repo.ListCountry(ctx, countryCode))
}

func (s *ServiceImpl) validateUserForCreate(user *model.User) error {
	if err := s.validateUser(user); err != nil {
		return err
	}
	if user.ID != "" {
		return fmt.Errorf("%w: id is filled by the service and shouldn't be filled", ErrInvalidParams)
	}
	if !user.CreatedAt.IsZero() {
		return fmt.Errorf("%w: created_at is filled by the service and shouldn't be filled", ErrInvalidParams)
	}
	if !user.UpdatedAt.IsZero() {
		return fmt.Errorf("%w: updated_at is filled by the service and shouldn't be filled", ErrInvalidParams)
	}
	if len(user.Password) < 8 {
		return fmt.Errorf("%w: password should be set and have at least 8 characters. %d proviced", ErrInvalidParams, len(user.Password))
	}
	return nil
}

func (s *ServiceImpl) validateUserForUpdate(user *model.User) error {
	if err := s.validateUser(user); err != nil {
		return err
	}
	if user.CreatedAt.IsZero() {
		return fmt.Errorf("%w: created_at should be provided", ErrInvalidParams)
	}
	if user.UpdatedAt.IsZero() {
		return fmt.Errorf("%w: updated_at should be provided", ErrInvalidParams)
	}
	if len(user.Password) > 0 && len(user.Password) < 8 {
		return fmt.Errorf("%w: if provided, password have at least 8 characters. %d proviced", ErrInvalidParams, len(user.Password))
	}
	return nil
}

func (s *ServiceImpl) validateUser(user *model.User) error {
	if len(user.FirstName) == 0 || len(user.FirstName) > 255 {
		return fmt.Errorf("%w: first_name should have between 1 and 255 characters, provided %d", ErrInvalidParams, len(user.FirstName))
	}
	if len(user.LastName) == 0 || len(user.LastName) > 255 {
		return fmt.Errorf("%w: last_name should have between 1 and 255 characters, provided %d", ErrInvalidParams, len(user.LastName))
	}
	if len(user.Name) == 0 || len(user.Name) > 255 {
		return fmt.Errorf("%w: name should have between 1 and 255 characters, provided %d", ErrInvalidParams, len(user.Name))
	}
	if len(user.Email) == 0 || len(user.Email) > 255 {
		return fmt.Errorf("%w: email should have between 1 and 255 characters, provided %d", ErrInvalidParams, len(user.Email))
	}
	if len(user.Country) != 2 {
		return fmt.Errorf("%w: country should have exactly 2 characters, got %d", ErrInvalidParams, len(user.Country))
	}
	if len(user.PasswordHash) > 0 || len(user.PasswordSalt) > 0 {
		return fmt.Errorf("%w: password_hash and password_salt should be empty as they're set by the service", ErrInvalidParams)
	}
	return nil
}

func (s *ServiceImpl) replacePasswordByHash(user *model.User) {
	user.PasswordSalt = randomSalt()
	salted := user.Password + user.PasswordSalt
	hashArray := sha256.Sum256([]byte(salted))
	user.PasswordHash = hex.EncodeToString(hashArray[:])
	user.Password = ""
}

func (s *ServiceImpl) removePasswords(users []*model.User, err error) ([]*model.User, error) {
	if err != nil {
		return nil, err
	}
	for _, u := range users {
		u.PasswordHash = ""
		u.PasswordSalt = ""
	}
	return users, nil
}

var timeNow = func() time.Time {
	return time.Now()
}

var uuidv1 = func() string {
	return uuid.Must(uuid.NewUUID()).String()
}

var randomSalt = func() string {
	saltBytes := make([]byte, 16)
	if _, err := rand.Read(saltBytes); err != nil {
		// this should be enough reason to panic
		panic(fmt.Errorf("can't read random bytes: %w", err))
	}
	return hex.EncodeToString(saltBytes)
}
