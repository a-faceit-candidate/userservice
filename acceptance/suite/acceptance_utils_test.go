package suite

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/a-faceit-candidate/restuser"
	"github.com/stretchr/testify/require"
)

const (
	someFirstName = "John"
	someLastName  = "Doe"
	someName      = "john_doe"
	someEmail     = "john@faceit.com"
	someCountry   = "es"
	somePassword  = "password"
)

const (
	someOtherFirstName = "Jane"
	someOtherLastName  = "Nil"
	someOtherName      = "jane_nil"
	someOtherEmail     = "jane@faceit.com"
	someOtherCountry   = "fr"
	someOtherPassword  = "security"
)

func (s *acceptanceSuite) saltedPasswordHash(password, salt string) string {
	hash := sha256.Sum256([]byte(password + salt))
	return hex.EncodeToString(hash[:])
}

func (s *acceptanceSuite) assertIsRestClientError(err error, expectedStatusCode int) {
	s.T().Helper()
	if restError := (restuser.Error{}); errors.As(err, &restError) {
		s.Equal(expectedStatusCode, restError.StatusCode)
	} else {
		s.FailNowf("Unexpected error received", "Expected a restuser.Error, got %T: %s", err, err)
	}
}

func rfc3339ToTime(t *testing.T, rfc3339 string) time.Time {
	parsed, err := time.Parse(time.RFC3339, rfc3339)
	require.NoError(t, err)
	return parsed
}
