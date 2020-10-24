package suite

import (
	"context"
	"net/http"
	"time"

	"github.com/a-faceit-candidate/restuser"
)

func (s *acceptanceSuite) TestDeleteUser() {
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

		err = s.client.DeleteUser(ctx, created.ID)
		s.Require().NoError(err)

		_, err = s.client.GetUser(ctx, created.ID)
		s.Require().Error(err)

		s.assertIsRestClientError(err, http.StatusNotFound)

		select {
		case id := <-s.userDeletedMessages:
			s.Equal(created.ID, id)
		case <-ctx.Done():
			s.Fail("Timeout waiting for the deleted NSQ message")
		}
	})

	s.Run("not found", func() {
		err := s.client.DeleteUser(ctx, "not-found")
		s.Require().Error(err)
		s.assertIsRestClientError(err, http.StatusNotFound)
	})
}
