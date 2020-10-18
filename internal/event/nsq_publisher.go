package event

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/a-faceit-candidate/userservice/internal/model"
)

const (
	topicUserCreated = "user.created"
	topicUserUpdated = "user.updated"
	topicUserDeleted = "user.deleted"
)

// NSQPublisher implements the persistence.CRUDObserver notifying the IDs of the changed entities through NSQ
type NSQPublisher struct {
	nsq nsqProducer
}

//go:generate mockery -inpkg -testonly -case underscore -name nsqProducer
type nsqProducer interface {
	Publish(string, []byte) error
}

func NewNSQPublisher(producer nsqProducer) *NSQPublisher {
	return &NSQPublisher{
		nsq: producer,
	}
}

func (n *NSQPublisher) OnCreate(ctx context.Context, user *model.User) error {
	return n.publish(ctx, topicUserCreated, user.ID)
}

func (n *NSQPublisher) OnUpdate(ctx context.Context, user *model.User) error {
	return n.publish(ctx, topicUserUpdated, user.ID)
}

func (n *NSQPublisher) OnDelete(ctx context.Context, userID string) error {
	return n.publish(ctx, topicUserDeleted, userID)
}

func (n *NSQPublisher) publish(_ context.Context, topic, userID string) error {
	data, err := json.Marshal(userID)
	if err != nil {
		return fmt.Errorf("can't marshal user ID: %s", err)
	}
	return n.nsq.Publish(topic, data)
}
