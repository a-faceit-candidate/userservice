package persistence

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/a-faceit-candidate/userservice/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	someUserID  = "foo"
	someUser    = &model.User{ID: someUserID}
	expectedErr = errors.New("the expected error")
)

func TestObservedRepository_Create(t *testing.T) {
	t.Run("repository call succeeds", func(t *testing.T) {
		repository := &MockRepository{}
		repository.On("Create", mock.Anything, someUser).Return(nil)

		observer1 := &MockCRUDObserver{}
		observer1.On("OnCreate", mock.Anything, someUser).Return(errors.New("broken"))

		observer2 := &MockCRUDObserver{}
		observer2.On("OnCreate", mock.Anything, someUser).Return(errors.New("broken too"))

		observed := NewObservedRepository(repository, observer1, observer2)
		err := observed.Create(context.Background(), someUser)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, observer1, observer2)
	})

	t.Run("repository call fails", func(t *testing.T) {
		repository := &MockRepository{}
		repository.On("Create", mock.Anything, someUser).Return(expectedErr)

		observer := &MockCRUDObserver{}

		observed := NewObservedRepository(repository, observer)
		err := observed.Create(context.Background(), someUser)
		assert.Equal(t, expectedErr, err)

		observer.AssertNotCalled(t, "OnCreate", mock.Anything, mock.Anything)
	})
}

func TestObservedRepository_Update(t *testing.T) {
	var someTime = time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)

	t.Run("repository call succeeds", func(t *testing.T) {
		repository := &MockRepository{}
		repository.On("Update", mock.Anything, someUser, someTime).Return(nil)

		observer1 := &MockCRUDObserver{}
		observer1.On("OnUpdate", mock.Anything, someUser).Return(errors.New("broken"))

		observer2 := &MockCRUDObserver{}
		observer2.On("OnUpdate", mock.Anything, someUser).Return(errors.New("broken too"))

		observed := NewObservedRepository(repository, observer1, observer2)
		err := observed.Update(context.Background(), someUser, someTime)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, observer1, observer2)
	})

	t.Run("repository call fails", func(t *testing.T) {
		repository := &MockRepository{}
		repository.On("Update", mock.Anything, someUser, someTime).Return(expectedErr)

		observer := &MockCRUDObserver{}

		observed := NewObservedRepository(repository, observer)
		err := observed.Update(context.Background(), someUser, someTime)
		assert.Equal(t, expectedErr, err)

		observer.AssertNotCalled(t, "OnUpdate", mock.Anything, mock.Anything)
	})
}

func TestObservedRepository_Delete(t *testing.T) {
	t.Run("repository call succeeds", func(t *testing.T) {
		repository := &MockRepository{}
		repository.On("Delete", mock.Anything, someUserID).Return(nil)

		observer1 := &MockCRUDObserver{}
		observer1.On("OnDelete", mock.Anything, someUserID).Return(errors.New("broken"))

		observer2 := &MockCRUDObserver{}
		observer2.On("OnDelete", mock.Anything, someUserID).Return(errors.New("broken too"))

		observed := NewObservedRepository(repository, observer1, observer2)
		err := observed.Delete(context.Background(), someUserID)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, observer1, observer2)
	})

	t.Run("repository call fails", func(t *testing.T) {
		repository := &MockRepository{}
		repository.On("Delete", mock.Anything, someUserID).Return(expectedErr)

		observer := &MockCRUDObserver{}

		observed := NewObservedRepository(repository, observer)
		err := observed.Delete(context.Background(), someUserID)
		assert.Equal(t, expectedErr, err)

		observer.AssertNotCalled(t, "OnDelete", mock.Anything, mock.Anything)
	})
}

func TestObservedRepository_ListAll(t *testing.T) {
	users := []*model.User{someUser, someUser}
	repository := &MockRepository{}
	repository.On("ListAll", mock.Anything).Return(users, nil)

	observer := &MockCRUDObserver{}

	observed := NewObservedRepository(repository, observer)
	got, err := observed.ListAll(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, users, got)
}

func TestObservedRepository_ListCountry(t *testing.T) {
	users := []*model.User{someUser, someUser}
	repository := &MockRepository{}
	repository.On("ListCountry", mock.Anything, "xx").Return(users, nil)

	observer := &MockCRUDObserver{}

	observed := NewObservedRepository(repository, observer)
	got, err := observed.ListCountry(context.Background(), "xx")
	assert.NoError(t, err)
	assert.Equal(t, users, got)
}
