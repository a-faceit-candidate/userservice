package suite

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/a-faceit-candidate/restuser"
)

func (s *acceptanceSuite) TestUpdateUser() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.Run("happy case", func() {
		// we use our own API for creating users as it's more maintainable than seeding the table
		created, err := s.client.CreateUser(ctx, &restuser.User{
			Name:    "john doe",
			Email:   "john@faceit.com",
			Country: "es",
		})
		s.Require().NoError(err)

		created.Name = someOtherName
		created.Email = someOtherEmail
		created.Country = someOtherCountry

		updated, err := s.client.UpdateUser(ctx, created)
		s.Require().NoError(err)
		s.Equal(created.ID, updated.ID)
		s.Equal(created.CreatedAt, updated.CreatedAt)
		s.NotEqual(created.UpdatedAt, updated.UpdatedAt)
		s.True(rfc3339ToTime(s.T(), updated.UpdatedAt).After(rfc3339ToTime(s.T(), created.UpdatedAt)), "Updated UpdatedAt shoudl be after creaated UpdatedAt")
		s.Equal(someOtherName, updated.Name)
		s.Equal(someOtherEmail, updated.Email)
		s.Equal(someOtherCountry, updated.Country)

		got, err := s.client.GetUser(ctx, created.ID)
		s.Require().NoError(err)
		s.Equal(created.ID, got.ID)
		s.Equal(created.CreatedAt, got.CreatedAt)
		s.Equal(updated.UpdatedAt, got.UpdatedAt)
		s.Equal(someOtherName, got.Name)
		s.Equal(someOtherEmail, got.Email)
		s.Equal(someOtherCountry, got.Country)

		select {
		case id := <-s.userUpdatedMessages:
			s.Equal(created.ID, id)
		case <-ctx.Done():
			s.Fail("Timeout waiting for the updated NSQ message")
		}
	})

	s.Run("conflict", func() {
		created, err := s.client.CreateUser(ctx, &restuser.User{
			Name:    "conflicting user",
			Email:   "conflict@faceit.com",
			Country: "de",
		})
		s.Require().NoError(err)

		// force an update, don't change anything
		_, err = s.client.UpdateUser(ctx, created)
		s.Require().NoError(err)

		// second update fails because we still use the UpdatedAt from the first model
		// for this to succeed we'd have use the one from the update user response
		_, err = s.client.UpdateUser(ctx, created)
		s.Require().Error(err)
		s.assertIsRestClientError(err, http.StatusConflict)
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
				created, err := s.client.CreateUser(ctx, &restuser.User{
					Name:    "won't be updated",
					Email:   "csgo@faceit.com",
					Country: "fi",
				})
				s.Require().NoError(err)

				created.Name = user.Name
				created.Email = user.Email
				created.Country = user.Country

				_, err = s.client.UpdateUser(ctx, created)
				s.Require().Error(err)
				s.assertIsRestClientError(err, http.StatusBadRequest)
			})
		}
	})

	s.Run("not found", func() {
		_, err := s.client.UpdateUser(ctx, &restuser.User{
			ID:      "not-found",
			Name:    "this won't be set",
			Email:   "whocares@faceit.com",
			Country: "ru",
		})
		s.Require().Error(err)
		s.assertIsRestClientError(err, http.StatusNotFound)
	})
}
