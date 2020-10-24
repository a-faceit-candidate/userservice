package suite

import (
	"context"
	"time"

	"github.com/a-faceit-candidate/restuser"
)

func (s *acceptanceSuite) TestListUsers() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	_, err := s.client.CreateUser(ctx, &restuser.User{
		FirstName: someFirstName,
		LastName:  someLastName,
		Name:      someName,
		Email:     someEmail,
		Password:  somePassword,
		Country:   someCountry,
	})
	s.Require().NoError(err)

	_, err = s.client.CreateUser(ctx, &restuser.User{
		FirstName: someOtherFirstName,
		LastName:  someOtherLastName,
		Name:      someOtherName,
		Email:     someOtherEmail,
		Password:  someOtherPassword,
		Country:   someOtherCountry,
	})
	s.Require().NoError(err)

	s.Run("all", func() {
		all, err := s.client.ListUsers(ctx, restuser.ListUsersParams{})
		s.Require().NoError(err)
		s.Require().Len(all, 2)
		s.Equal(someName, all[0].Name)
		s.Equal(someOtherName, all[1].Name)
	})

	s.Run("by country", func() {
		all, err := s.client.ListUsers(ctx, restuser.ListUsersParams{Country: someOtherCountry})
		s.Require().NoError(err)
		s.Require().Len(all, 1)
		s.Equal(someOtherName, all[0].Name)
	})
}
