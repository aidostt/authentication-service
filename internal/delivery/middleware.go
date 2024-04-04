package delivery

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	authorizationHeader = "Authorization"

	userCtx = "userId"
)

func (h *Handler) userIdentity(c *gin.Context) {
	id, err := h.parseAuthHeader(c)
	if err != nil {
		fmt.Println("i was here before you")
		if err.Error() == "token is expired" {
			fmt.Println("i was here")
			// Token is expired, attempt to refresh it
			newAT, refreshErr := h.services.Session.RefreshTokens(c.Request.Context(), c.GetHeader(authorizationHeader))
			if refreshErr != nil {
				newResponse(c, http.StatusUnauthorized, refreshErr.Error())
				return
			}

			// Set the new token as a cookie
			c.SetCookie("jwt", newAT, 3600, "/", "localhost", false, true)

			// Update the context with the new user ID
			id, _ = h.parseAuthHeader(c)
		} else {
			newResponse(c, http.StatusUnauthorized, err.Error())
			return
		}
	}

	c.Set(userCtx, id)
}

func (h *Handler) parseAuthHeader(c *gin.Context) (string, error) {
	token, err := c.Cookie("jwt")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return "", errors.New("unauthorized access")
		}
		return "", err
	}

	return h.tokenManager.Parse(token)
}

func corsMiddleware(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "*")
	c.Header("Access-Control-Allow-Headers", "*")
	c.Header("Content-Type", "application/json")

	if c.Request.Method != "OPTIONS" {
		c.Next()
	} else {
		c.AbortWithStatus(http.StatusOK)
	}
}
