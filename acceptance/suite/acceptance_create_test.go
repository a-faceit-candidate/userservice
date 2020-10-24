package suite

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/a-faceit-candidate/restuser"
)

func (s *acceptanceSuite) TestCreateUser() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.Run("happy case", func() {
		testStart := time.Now()

		created, err := s.client.CreateUser(ctx, &restuser.User{
			FirstName: someFirstName,
			LastName:  someLastName,
			Name:      someName,
			Email:     someEmail,
			Password:  somePassword,
			Country:   someCountry,
		})
		s.Require().NoError(err)

		s.Require().NotEmpty(created.ID)

		s.Equal(someFirstName, created.FirstName)
		s.Equal(someLastName, created.LastName)
		s.Equal(someName, created.Name)
		s.Equal(someEmail, created.Email)
		s.Equal(someCountry, created.Country)
		s.Empty(created.Password)
		s.NotEmpty(created.PasswordSalt)
		s.Equal(s.saltedPasswordHash(somePassword, created.PasswordSalt), created.PasswordHash)

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
			"id provided": {
				ID:        "injected",
				FirstName: someFirstName,
				LastName:  someLastName,
				Name:      someName,
				Email:     someEmail,
				Password:  somePassword,
				Country:   someCountry,
			},
			"created_at provided": {
				CreatedAt: time.Now().Format(time.RFC3339Nano),
				FirstName: someFirstName,
				LastName:  someLastName,
				Name:      someName,
				Email:     someEmail,
				Password:  somePassword,
				Country:   someCountry,
			},
			"updated_at provided": {
				UpdatedAt: time.Now().Format(time.RFC3339Nano),
				FirstName: someFirstName,
				LastName:  someLastName,
				Name:      someName,
				Email:     someEmail,
				Password:  somePassword,
				Country:   someCountry,
			},
			"password hash provided": {
				FirstName:    someFirstName,
				LastName:     someLastName,
				Name:         someName,
				Email:        someEmail,
				Password:     somePassword,
				Country:      someCountry,
				PasswordHash: "injected",
			},
			"password salt provided": {
				FirstName:    someFirstName,
				LastName:     someLastName,
				Name:         someName,
				Email:        someEmail,
				Password:     somePassword,
				Country:      someCountry,
				PasswordSalt: "injected",
			},
			"empty first name": {
				LastName: someLastName,
				Name:     someName,
				Email:    someEmail,
				Password: somePassword,
				Country:  someCountry,
			},
			"empty last name": {
				FirstName: someFirstName,
				Name:      someName,
				Email:     someEmail,
				Password:  somePassword,
				Country:   someCountry,
			},
			"empty name": {
				FirstName: someFirstName,
				LastName:  someLastName,
				Email:     someEmail,
				Password:  somePassword,
				Country:   someCountry,
			},
			"too long name": {
				FirstName: someFirstName,
				LastName:  someLastName,
				Name:      strings.Repeat("f", 256),
				Email:     someEmail,
				Password:  somePassword,
				Country:   someCountry,
			},
			"empty email": {
				FirstName: someFirstName,
				LastName:  someLastName,
				Name:      someName,
				Password:  somePassword,
				Country:   someCountry,
			},
			"empty password": {
				FirstName: someFirstName,
				LastName:  someLastName,
				Name:      someName,
				Email:     someEmail,
				Country:   someCountry,
			},
			"too short password": {
				FirstName: someFirstName,
				LastName:  someLastName,
				Name:      someName,
				Email:     someEmail,
				Password:  "1234567",
				Country:   someCountry,
			},
			"non two character country": {
				FirstName: someFirstName,
				LastName:  someLastName,
				Name:      someName,
				Email:     someEmail,
				Password:  somePassword,
				Country:   "xxx",
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
