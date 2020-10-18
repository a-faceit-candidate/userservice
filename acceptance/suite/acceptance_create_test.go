package suite

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/a-faceit-candidate/restuser"
	"github.com/stretchr/testify/require"
)

const (
	someName    = "john doe"
	someEmail   = "john@faceit.com"
	someCountry = "es"
)

const (
	someOtherName    = "jane doe"
	someOtherEmail   = "jane@faceit.com"
	someOtherCountry = "fr"
)

func (s *acceptanceSuite) TestCreateUser() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.Run("happy case", func() {
		testStart := time.Now()

		created, err := s.client.CreateUser(ctx, &restuser.User{
			Name:    "john doe",
			Email:   "john@faceit.com",
			Country: "es",
		})
		s.Require().NoError(err)

		s.Require().NotEmpty(created.ID)
		s.Equal(someName, created.Name)
		s.Equal(someEmail, created.Email)
		s.Equal(someCountry, created.Country)

		s.True(rfc3339ToTime(s.T(), created.CreatedAt).After(testStart))
		s.True(rfc3339ToTime(s.T(), created.UpdatedAt).After(testStart))

		select {
		case id := <-s.userCreatedMessages:
			s.Equal(created.ID, id)
		case <-ctx.Done():
			s.Fail("Timeout waiting for the created NSQ message")
		}
	})

	s.Run("invalid params", func() {
		for testName, user := range map[string]*restuser.User{
			"empty name": {
				Email:   "bar",
				Country: "uk",
			},
			"too long name": {
				Name:    strings.Repeat("f", 256),
				Email:   "bar",
				Country: "uk",
			},
			"empty email": {
				Name:    "foo",
				Country: "uk",
			},
			"non two character country": {
				Name:    "foo",
				Email:   "bar",
				Country: "zzz",
			},
		} {
			s.Run(testName, func() {
				_, err := s.client.CreateUser(ctx, user)
				s.Require().Error(err)
				s.assertIsRestClientError(err, http.StatusBadRequest)
			})
		}
	})
}

func (s *acceptanceSuite) assertIsRestClientError(err error, expectedStatusCode int) {
	s.T().Helper()
	if restError := (restuser.Error{}); errors.As(err, &restError) {
		s.Equal(expectedStatusCode, restError.StatusCode)
	} else {
		s.FailNowf("Unexpected error received", "Expected a restuser.Error, got %T", err)
	}
}

func rfc3339ToTime(t *testing.T, rfc3339 string) time.Time {
	parsed, err := time.Parse(time.RFC3339, rfc3339)
	require.NoError(t, err)
	return parsed
}
