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
			FirstName: someFirstName,
			LastName:  someLastName,
			Name:      someName,
			Email:     someEmail,
			Password:  somePassword,
			Country:   someCountry,
		})
		s.Require().NoError(err)

		created.FirstName = someOtherFirstName
		created.LastName = someOtherLastName
		created.Name = someOtherName
		created.Email = someOtherEmail
		created.Country = someOtherCountry
		created.Password = someOtherPassword
		created.PasswordHash = ""
		created.PasswordSalt = ""

		updated, err := s.client.UpdateUser(ctx, created)
		s.Require().NoError(err)
		s.Equal(created.ID, updated.ID)
		s.Equal(created.CreatedAt, updated.CreatedAt)
		s.NotEqual(created.UpdatedAt, updated.UpdatedAt)
		s.True(rfc3339ToTime(s.T(), updated.UpdatedAt).After(rfc3339ToTime(s.T(), created.UpdatedAt)), "Updated UpdatedAt shoudl be after creaated UpdatedAt")
		s.Equal(someOtherFirstName, updated.FirstName)
		s.Equal(someOtherLastName, updated.LastName)
		s.Equal(someOtherName, updated.Name)
		s.Equal(someOtherEmail, updated.Email)
		s.Equal(someOtherCountry, updated.Country)
		s.Empty(updated.Password)
		s.NotEmpty(updated.PasswordSalt)
		s.Equal(s.saltedPasswordHash(someOtherPassword, updated.PasswordSalt), updated.PasswordHash)

		got, err := s.client.GetUser(ctx, created.ID)
		s.Require().NoError(err)
		s.Equal(created.ID, got.ID)
		s.Equal(created.CreatedAt, got.CreatedAt)
		s.Equal(updated.UpdatedAt, got.UpdatedAt)
		s.Equal(someOtherFirstName, got.FirstName)
		s.Equal(someOtherLastName, got.LastName)
		s.Equal(someOtherName, got.Name)
		s.Equal(someOtherEmail, got.Email)
		s.Equal(someOtherCountry, got.Country)
		s.Empty(got.Password)
		s.NotEmpty(got.PasswordSalt)
		s.Equal(s.saltedPasswordHash(someOtherPassword, got.PasswordSalt), got.PasswordHash)

		select {
		case id := <-s.userUpdatedMessages:
			s.Equal(created.ID, id)
		case <-ctx.Done():
			s.Fail("Timeout waiting for the updated NSQ message")
		}
	})

	s.Run("no password update", func() {
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

		hash := created.PasswordHash
		salt := created.PasswordSalt

		created.FirstName = someOtherFirstName
		created.PasswordHash = ""
		created.PasswordSalt = ""

		_, err = s.client.UpdateUser(ctx, created)
		s.Require().NoError(err)

		got, err := s.client.GetUser(ctx, created.ID)
		s.Require().NoError(err)
		s.Equal(hash, got.PasswordHash)
		s.Equal(salt, got.PasswordSalt)
	})

	s.Run("conflict", func() {
		created, err := s.client.CreateUser(ctx, &restuser.User{
			FirstName: someFirstName,
			LastName:  someLastName,
			Name:      someName,
			Email:     someEmail,
			Password:  somePassword,
			Country:   someCountry,
		})
		s.Require().NoError(err)

		created.PasswordHash = ""
		created.PasswordSalt = ""

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
				Country:   "uk",
			},
			"non two character country": {
				FirstName: someFirstName,
				LastName:  someLastName,
				Name:      someName,
				Email:     someEmail,
				Password:  somePassword,
				Country:   "zzz",
			},
			"password hash provided": {
				FirstName:    someFirstName,
				LastName:     someLastName,
				Name:         someName,
				Email:        someEmail,
				Password:     somePassword,
				PasswordHash: "injected",
				Country:      someCountry,
			},
			"password salt provided": {
				FirstName:    someFirstName,
				LastName:     someLastName,
				Name:         someName,
				Email:        someEmail,
				Password:     somePassword,
				PasswordSalt: "injected",
				Country:      someCountry,
			},
		} {
			s.Run(testName, func() {
				toUpdate, err := s.client.CreateUser(ctx, &restuser.User{
					FirstName: someFirstName,
					LastName:  someLastName,
					Name:      someName,
					Email:     someEmail,
					Password:  somePassword,
					Country:   someCountry,
				})
				s.Require().NoError(err)

				toUpdate.FirstName = user.FirstName
				toUpdate.LastName = user.LastName
				toUpdate.Name = user.Name
				toUpdate.Email = user.Email
				toUpdate.Country = user.Country
				toUpdate.Password = user.Password
				toUpdate.PasswordHash = user.PasswordHash
				toUpdate.PasswordSalt = user.PasswordSalt

				_, err = s.client.UpdateUser(ctx, toUpdate)
				s.Require().Error(err)
				s.assertIsRestClientError(err, http.StatusBadRequest)
			})
		}
	})

	s.Run("not found", func() {
		_, err := s.client.UpdateUser(ctx, &restuser.User{
			ID:        "this-user-does-not-exist",
			CreatedAt: time.Now().Format(time.RFC3339Nano),
			UpdatedAt: time.Now().Format(time.RFC3339Nano),
			FirstName: someFirstName,
			LastName:  someLastName,
			Name:      someName,
			Email:     someEmail,
			Password:  somePassword,
			Country:   someCountry,
		})
		s.Require().Error(err)
		s.assertIsRestClientError(err, http.StatusNotFound)
	})
}
