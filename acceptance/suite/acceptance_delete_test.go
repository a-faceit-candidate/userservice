package suite

import (
	"context"
	"errors"
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
			Name:    "john doe",
			Email:   "john@faceit.com",
			Country: "es",
		})
		s.Require().NoError(err)

		created.Name = someOtherName
		created.Email = someOtherEmail
		created.Country = someOtherCountry

		err = s.client.DeleteUser(ctx, created.ID)
		s.Require().NoError(err)

		_, err = s.client.GetUser(ctx, created.ID)
		s.Require().Error(err)

		if restError := (restuser.Error{}); errors.As(err, &restError) {
			s.Equal(http.StatusNotFound, restError.StatusCode)
		} else {
			s.FailNowf("Unexpected error received", "Expected a restuser.Error, got %T", err)
		}

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
