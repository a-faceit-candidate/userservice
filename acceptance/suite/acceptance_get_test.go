package suite

import (
	"context"
	"net/http"
	"time"

	"github.com/a-faceit-candidate/restuser"
)

func (s *acceptanceSuite) TestGetUser() {
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

		got, err := s.client.GetUser(ctx, created.ID)
		s.Require().NoError(err)
		s.Equal(created.ID, got.ID)
		s.Equal(created.CreatedAt, got.CreatedAt)
		s.Equal(created.UpdatedAt, got.UpdatedAt)
		s.Equal(someName, got.Name)
		s.Equal(someEmail, got.Email)
		s.Equal(someCountry, got.Country)
	})

	s.Run("not found", func() {
		_, err := s.client.GetUser(ctx, "not-found")
		s.Require().Error(err)
		s.assertIsRestClientError(err, http.StatusNotFound)
	})
}
