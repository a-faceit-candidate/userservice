package persistence

import (
	"context"
	"time"

	"github.com/a-faceit-candidate/userservice/internal/log"
	"github.com/a-faceit-candidate/userservice/internal/model"
)

// CRUDObserver defines the repository operations observer functionality.
type CRUDObserver interface {
	OnCreate(context.Context, *model.User) error
	OnUpdate(context.Context, *model.User) error
	OnDelete(context.Context, string) error
}

//go:generate mockery -inpkg -testonly -case underscore -name CRUDObserver

// ObservedRepository will try to notify the registered observers on CRUD operations.
// Failed observer calls won't fail the CRUD calls on this repository implementation, they will be logged as warnings.
type ObservedRepository struct {
	Repository
	observers []CRUDObserver
}

// NewObservedRepository creates a new ObservedRepository
func NewObservedRepository(repo Repository, observers ...CRUDObserver) *ObservedRepository {
	return &ObservedRepository{
		Repository: repo,
		observers:  observers,
	}
}

func (r *ObservedRepository) Create(ctx context.Context, user *model.User) error {
	err := r.Repository.Create(ctx, user)
	if err != nil {
		return err
	}
	for _, ob := range r.observers {
		if err := ob.OnCreate(ctx, user); err != nil {
			log.For(ctx).Warningf("Can't notify observer %T OnCreate: %s", ob, err)
		}
	}
	return nil
}

func (r *ObservedRepository) Update(ctx context.Context, user *model.User, prevUpdatedAt time.Time) error {
	err := r.Repository.Update(ctx, user, prevUpdatedAt)
	if err != nil {
		return err
	}
	for _, ob := range r.observers {
		if err := ob.OnUpdate(ctx, user); err != nil {
			log.For(ctx).Warningf("Can't notify observer %T OnUpdate: %s", ob, err)
		}
	}
	return nil
}

func (r *ObservedRepository) Delete(ctx context.Context, id string) error {
	err := r.Repository.Delete(ctx, id)
	if err != nil {
		return err
	}
	for _, ob := range r.observers {
		if err := ob.OnDelete(ctx, id); err != nil {
			log.For(ctx).Warningf("Can't notify observer %T OnDelete: %s", ob, err)
		}
	}
	return nil
}
