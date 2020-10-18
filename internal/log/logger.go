package log

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// For creates a logger (logrus.Entry) with fields from the context
func For(ctx context.Context) *logrus.Entry {
	fields, ok := ctx.Value(baggageKey).(map[string]interface{})
	if !ok {
		return logrus.NewEntry(logrus.StandardLogger())
	}
	return logrus.WithFields(fields)
}

type baggageType int

var baggageKey baggageType

// WithValues returns a context with all key value pairs added to the baggage store.
func WithValues(ctx context.Context, keyValue map[string]interface{}) context.Context {
	oldBaggage, ok := ctx.Value(baggageKey).(map[string]interface{})
	if !ok {
		return context.WithValue(ctx, baggageKey, map[string]interface{}(keyValue))
	}

	newBaggage := make(map[string]interface{}, len(oldBaggage)+len(keyValue))
	for oldbaggageKey, oldValue := range oldBaggage {
		newBaggage[oldbaggageKey] = oldValue
	}

	for newbaggageKey, newValue := range keyValue {
		newBaggage[newbaggageKey] = newValue
	}

	return context.WithValue(ctx, baggageKey, newBaggage)
}

// AddLogContextBaggage is a gin middleware to add the request information to the logging baggage in the context
func AddLogContextBaggage(c *gin.Context) {
	ctx := WithValues(c.Request.Context(), map[string]interface{}{
		"method": c.Request.Method,
		"uri":    c.Request.RequestURI,
	})
	c.Request = c.Request.WithContext(ctx)
	c.Next()
	// we could log here all the requests, but we don't want to be so noisy
}
