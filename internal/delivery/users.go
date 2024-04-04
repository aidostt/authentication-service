package delivery

import (
	"authentication-service/internal/domain"
	"authentication-service/internal/service"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func (h *Handler) initUsersRoutes(api *gin.RouterGroup) {
	users := api.Group("/users")
	{
		users.POST("/sign-up", h.userSignUp)
		users.POST("/sign-in", h.userSignIn)
		users.POST("/auth/refresh", h.userRefresh)
		authenticated := users.Group("/", h.userIdentity)
		{
			authenticated.GET("/healthcheck", h.healthcheck)
			users.POST("/sign-out", h.logout)

		}
	}
}

func (h *Handler) userSignUp(c *gin.Context) {
	var inp userSignUpInput
	if err := c.BindJSON(&inp); err != nil {
		newResponse(c, http.StatusBadRequest, "invalid input body")

		return
	}
	AT, err := h.services.Users.SignUp(c.Request.Context(), service.UserSignUpInput{
		Email:    inp.Email,
		Password: inp.Password,
	})
	if err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			newResponse(c, http.StatusBadRequest, err.Error())
			return
		}
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.SetCookie("jwt", AT, time.Now().Second()+900, "/", "", false, true)
	c.JSON(http.StatusOK, tokenResponse{AccessToken: AT})
	c.Status(http.StatusCreated)
}

func (h *Handler) userSignIn(c *gin.Context) {
	var inp signInInput
	if err := c.BindJSON(&inp); err != nil {
		newResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	AT, err := h.services.Users.SignIn(c.Request.Context(), service.UserSignInInput{
		Email:    inp.Email,
		Password: inp.Password,
	})
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			newResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.SetCookie("jwt", AT, time.Now().Second()+900, "/", "", false, true)
	c.JSON(http.StatusOK, tokenResponse{AccessToken: AT})
	//c.SetCookie("refresh", res.RefreshToken, time.Now().Second()+3600, "/", "", false, true)
}

func (h *Handler) userRefresh(c *gin.Context) {
	token, err := c.Cookie("jwt")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			newResponse(c, http.StatusUnauthorized, "unauthorized access")
			return
		}
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	AT, err := h.services.Session.RefreshTokens(c.Request.Context(), token)
	if err != nil {
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.SetCookie("jwt", AT, time.Now().Second()+900, "/", "", false, true)
	c.JSON(http.StatusOK, tokenResponse{AccessToken: AT})
}

func (h *Handler) logout(c *gin.Context) {
	c.SetCookie("jwt", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, healthResponse{
		Status: "success",
	})
}
