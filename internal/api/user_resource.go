package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/a-faceit-candidate/restuser"
	"github.com/a-faceit-candidate/userservice/internal/log"
	"github.com/a-faceit-candidate/userservice/internal/model"
	"github.com/a-faceit-candidate/userservice/internal/service"
	"github.com/gin-gonic/gin"
)

// httpStatusRequestCanceled is not a real status code, but one copied from nginx implementation
// we return this when the request context is canceled, i.e. the client has closed the http connection
// since the client isn't listening anymore, we they don't care about the status code we send, however
// we don't want to map those context.Canceled error to 5xx as they're a client-side error.
const httpStatusRequestCanceled = 499

// UsersResource handles /users resource
type UsersResource struct {
	svc service.Service
}

func NewUsersResource(svc service.Service) *UsersResource {
	return &UsersResource{
		svc: svc,
	}
}

func (res *UsersResource) AddRoutes(r gin.IRouter) {
	base := r.Group("/users")
	base.GET("/", res.get)
	base.GET("/:id", res.getByID)
	base.DELETE("/:id", res.deleteByID)
	base.PUT("/:id", res.putByID)
	base.POST("/", res.post)
}

func (res *UsersResource) post(c *gin.Context) {
	ctx := c.Request.Context()
	ru := &restuser.User{}
	if err := c.BindJSON(ru); err != nil {
		log.For(ctx).Infof("Received a malformed payload: %s", err)
		c.JSON(http.StatusBadRequest, errorResponse("Can't bind request payload: %s", err))
		return
	}

	user, err := restToUser(ru)
	if err != nil {
		log.For(ctx).Infof("Can't map rest model to internal: %s", err)
		c.JSON(http.StatusBadRequest, errorResponse("Can't map model: %s", err))
		return
	}

	user, err = res.svc.Create(ctx, user)
	if err != nil {
		res.handleError(ctx, c, err)
		return
	}

	ctx = log.WithValues(ctx, map[string]interface{}{"user_id": user.ID})
	log.For(ctx).Info("Successfully created user")

	c.JSON(http.StatusCreated, userToREST(user))
}

func (res *UsersResource) getByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	ctx = log.WithValues(ctx, map[string]interface{}{"user_id": id})

	user, err := res.svc.Get(ctx, id)
	if err != nil {
		res.handleError(ctx, c, err)
		return
	}

	c.JSON(http.StatusOK, userToREST(user))
}

func (res *UsersResource) deleteByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	ctx = log.WithValues(ctx, map[string]interface{}{"user_id": id})

	err := res.svc.Delete(ctx, id)
	if err != nil {
		res.handleError(ctx, c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (res *UsersResource) putByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	ctx = log.WithValues(ctx, map[string]interface{}{"user_id": id})

	ru := &restuser.User{}
	if err := c.BindJSON(ru); err != nil {
		log.For(ctx).Infof("Received a malformed payload: %s", err)
		c.JSON(http.StatusBadRequest, errorResponse("Can't bind request payload: %s", err))
		return
	}

	user, err := restToUser(ru)
	if err != nil {
		log.For(ctx).Infof("Can't map rest model to internal: %s", err)
		c.JSON(http.StatusBadRequest, errorResponse("Can't map model: %s", err))
		return
	}

	user, err = res.svc.Update(ctx, id, user)
	if err != nil {
		res.handleError(ctx, c, err)
		return
	}

	log.For(ctx).Info("Successfully updated user")

	c.JSON(http.StatusOK, userToREST(user))
}

func (res *UsersResource) get(c *gin.Context) {
	ctx := c.Request.Context()

	var (
		users []*model.User
		err   error
	)

	country := c.Query("country")
	if country != "" {
		ctx = log.WithValues(ctx, map[string]interface{}{"country": country})
		users, err = res.svc.ListCountry(ctx, country)
	} else {
		users, err = res.svc.ListAll(ctx)
	}

	if err != nil {
		res.handleInternalError(ctx, c, err)
		return
	}

	restUsers := make([]restuser.User, len(users))
	for i, u := range users {
		restUsers[i] = userToREST(u)
	}

	c.JSON(http.StatusOK, restUsers)
}

func (res *UsersResource) handleError(ctx context.Context, c *gin.Context, err error) {
	if res.handleServiceError(ctx, c, err) {
		return
	}
	res.handleInternalError(ctx, c, err)
}

var serviceErrorToStatusCode = map[error]int{
	service.ErrNotFound:      http.StatusNotFound,
	service.ErrInvalidParams: http.StatusBadRequest,
	service.ErrConflict:      http.StatusConflict,
}

func (res *UsersResource) handleServiceError(_ context.Context, c *gin.Context, err error) bool {
	for svcErr, statusCode := range serviceErrorToStatusCode {
		if errors.Is(err, svcErr) {
			c.JSON(statusCode, errorResponse(err.Error()))
			return true
		}
	}
	return false
}

func (res *UsersResource) handleInternalError(ctx context.Context, c *gin.Context, err error) {
	if errors.Is(err, context.Canceled) {
		c.Status(httpStatusRequestCanceled)
		return
	}
	log.For(ctx).Errorf("Internal error: %s", err)
	c.JSON(http.StatusInternalServerError, errorResponse("Internal error: %s", err))
}

func restToUser(ru *restuser.User) (u *model.User, err error) {
	var createdAt time.Time
	if ru.CreatedAt != "" {
		createdAt, err = time.Parse(time.RFC3339, ru.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("can't parse created_at: %w", err)
		}
	}
	var updatedAt time.Time
	if ru.UpdatedAt != "" {
		updatedAt, err = time.Parse(time.RFC3339, ru.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("can't parse updated_at: %w", err)
		}
	}

	return &model.User{
		ID:           ru.ID,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
		FirstName:    ru.FirstName,
		LastName:     ru.LastName,
		Name:         ru.Name,
		Email:        ru.Email,
		Password:     ru.Password,
		PasswordHash: ru.PasswordHash,
		PasswordSalt: ru.PasswordSalt,
		Country:      ru.Country,
	}, nil
}

func userToREST(u *model.User) restuser.User {
	return restuser.User{
		ID:           u.ID,
		CreatedAt:    u.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:    u.UpdatedAt.Format(time.RFC3339Nano),
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		Name:         u.Name,
		Email:        u.Email,
		Password:     u.Password,
		PasswordHash: u.PasswordHash,
		PasswordSalt: u.PasswordSalt,
		Country:      u.Country,
	}
}

func errorResponse(msg string, args ...interface{}) restuser.ErrorResponse {
	return restuser.ErrorResponse{
		Message: fmt.Sprintf(msg, args...),
	}
}
