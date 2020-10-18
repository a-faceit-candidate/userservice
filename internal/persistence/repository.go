package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/a-faceit-candidate/userservice/internal/model"
)

type Repository interface {
	// Create will create a user. It expects the ID and CreatedAt, UpdatedAt fields to be filled.
	// It will fail with ErrConflict there's already a user with that ID.
	Create(context.Context, *model.User) error
	// Update will update the user with same ID and same prevUpdatedAt timestamp,
	// if the UpdatedAt in the DB differs, it fail with ErrConflict.
	// If user doesn't exist, it will fail with ErrNotFound
	Update(ctx context.Context, user *model.User, prevUpdatedAt time.Time) error
	// Get will retrieve a user with the ID provided, or ErrNotFound if not found.
	Get(context.Context, string) (*model.User, error)
	// Delete will delete the user with the ID provided. It will return ErrNotFound if no users were found.
	Delete(context.Context, string) error
	// ListAll retrieves all the users.
	ListAll(context.Context) ([]*model.User, error)
	// ListCountry retrieves all the users from the given country.
	ListCountry(ctx context.Context, countryCode string) ([]*model.User, error)
}

//go:generate mockery -output persistencemock -outpkg persistencemock -case unserscore -name Repository
//go:generate mockery -inpkg -testonly -case underscore -name Repository

var ErrNotFound = errors.New("not found")
var ErrConflict = errors.New("conflict updating")
