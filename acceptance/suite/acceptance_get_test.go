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
			FirstName: someFirstName,
			LastName:  someLastName,
			Name:      someName,
			Email:     someEmail,
			Password:  somePassword,
			Country:   someCountry,
		})
		s.Require().NoError(err)

		got, err := s.client.GetUser(ctx, created.ID)
		s.Require().NoError(err)
		s.Equal(created.ID, got.ID)
		s.Equal(created.CreatedAt, got.CreatedAt)
		s.Equal(created.UpdatedAt, got.UpdatedAt)
		s.Equal(someFirstName, got.FirstName)
		s.Equal(someLastName, got.LastName)
		s.Equal(someName, got.Name)
		s.Equal(someEmail, got.Email)
		s.Equal(someCountry, got.Country)
		s.Equal(created.PasswordHash, got.PasswordHash)
		s.Equal(created.PasswordSalt, got.PasswordSalt)
	})

	s.Run("not found", func() {
		_, err := s.client.GetUser(ctx, "not-found")
		s.Require().Error(err)
		s.assertIsRestClientError(err, http.StatusNotFound)
	})
}
