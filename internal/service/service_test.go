package service

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/a-faceit-candidate/userservice/internal/model"
	"github.com/a-faceit-candidate/userservice/internal/persistence/persistencemock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	mockedNow  = time.Date(2020, 1, 2, 3, 4, 5, int(time.Millisecond+2*time.Microsecond+3*time.Nanosecond), time.UTC)
	mockedUUID = "123e4567-e89b-12d3-a456-426614174000"
)

func TestMain(t *testing.M) {
	// mocking for poor people happens here.
	// we could also build a test suite
	// or maybe we could follow Uncle Bob's rules and inject these two things in the constructor
	// all are valid options, I chose this one
	timeNow = func() time.Time { return mockedNow }
	uuidv1 = func() string { return mockedUUID }
	os.Exit(t.Run())
}

func TestServiceImpl_Create(t *testing.T) {
	t.Run("happy case", func(t *testing.T) {
		user := model.User{
			Name:    "foo",
			Email:   "bar@hotmail.com",
			Country: "zz",
		}
		expectedRepositoryUser := user
		expectedRepositoryUser.ID = mockedUUID

		expectedRepositoryUser.CreatedAt = mockedNow.Truncate(time.Microsecond)
		expectedRepositoryUser.UpdatedAt = mockedNow.Truncate(time.Microsecond)

		repository := &persistencemock.Repository{}
		repository.On("Create", mock.Anything, &expectedRepositoryUser).Return(nil)

		svc := New(repository)

		got, err := svc.Create(context.Background(), &user)
		assert.NoError(t, err)
		assert.Equal(t, &expectedRepositoryUser, got)
	})

	t.Run("invalid user params", func(t *testing.T) {
		/*
			This should be a set of tests with invalid params, like the one we have in the acceptance tests.
			Since I didn't write them, I'll leave here a funny dolphin instead:
			                                  __
			                               _.-~  )
			                    _..--~~~~,'   ,-/     _
			                 .-'. . . .'   ,-','    ,' )
			               ,'. . . _   ,--~,-'__..-'  ,'
			             ,'. . .  (@)' ---~~~~      ,'
			            /. . . . '~~             ,-'
			           /. . . . .             ,-'
			          ; . . . .  - .        ,'
			         : . . . .       _     /
			        . . . . .          `-.:
			       . . . ./  - .          )
			      .  . . |  _____..---.._/ ____ Seal _
			~---~~~~----~~~~             ~~
		*/
	})

	t.Run("repository conflict", func(t *testing.T) {
		/*
				This should be testing that when repository returns persistence.ErrConflict, we return an internal error since it's our fault.
				I didn't write this test either, but I found this funny horse ascii art on the internet:

			                            _(\_/)
			                          ,((((^`\
			                         ((((  (6 \
			                       ,((((( ,    \
			   ,,,_              ,(((((  /"._  ,`,
			  ((((\\ ,...       ,((((   /    `-.-'
			  )))  ;'    `"'"'""((((   (
			 (((  /            (((      \
			  )) |                      |
			 ((  |        .       '     |
			 ))  \     _ '      `t   ,.')
			 (   |   y;- -,-""'"-.\   \/
			 )   / ./  ) /         `\  \
			    |./   ( (           / /'
			    ||     \\          //'|
			jgs ||      \\       _//'||
			    ||       ))     |_/  ||
			    \_\     |_/          ||
			    `'"                  \_\
		*/
	})
}

/*
Now hundreds of lines of testcases should follow, but let's be honest, nobody would read them all when reviewing
a code challenge, so instead we can have some fun time and look at this elephant by Joan G. Stark:
								_
							  .' `'.__
							 /      \ `'"-,
			.-''''--...__..-/ .     |      \
		  .'               ; :'     '.  a   |
		 /                 | :.       \     =\
		;                   \':.      /  ,-.__;.-;`
	   /|     .              '--._   /-.7`._..-;`
	  ; |       '                |`-'      \  =|
	  |/\        .   -' /     /  ;         |  =/
	  (( ;.       ,_  .:|     | /     /\   | =|
	   ) / `\     | `""`;     / |    | /   / =/
		 | ::|    |      \    \ \    \ `--' =/
		/  '/\    /       )    |/     `-...-`
	   /    | |  `\    /-'    /;
	   \  ,,/ |    \   D    .'  \
	jgs `""`   \  nnh  D_.-'L__nnh
				`"""`
*/
