package event

import (
	"context"
	"errors"
	"testing"

	"github.com/a-faceit-candidate/userservice/internal/model"
	"github.com/go-playground/assert/v2"
)

const someUserID = "asdf-asdf"

var someUserIDJSON = []byte(`"asdf-asdf"`)

var expectedErr = errors.New("the expected error")

func TestNSQPublisher_OnCreate(t *testing.T) {
	producer := &mockNsqProducer{}
	producer.On("Publish", topicUserCreated, someUserIDJSON).Return(expectedErr)

	publisher := NewNSQPublisher(producer)
	err := publisher.OnCreate(context.Background(), &model.User{ID: someUserID})
	assert.Equal(t, expectedErr, err)
}

func TestNSQPublisher_OnUpdate(t *testing.T) {
	producer := &mockNsqProducer{}
	producer.On("Publish", topicUserUpdated, someUserIDJSON).Return(expectedErr)

	publisher := NewNSQPublisher(producer)
	err := publisher.OnUpdate(context.Background(), &model.User{ID: someUserID})
	assert.Equal(t, expectedErr, err)
}

func TestNSQPublisher_OnDelete(t *testing.T) {
	producer := &mockNsqProducer{}
	producer.On("Publish", topicUserDeleted, someUserIDJSON).Return(expectedErr)

	publisher := NewNSQPublisher(producer)
	err := publisher.OnDelete(context.Background(), someUserID)
	assert.Equal(t, expectedErr, err)
}
