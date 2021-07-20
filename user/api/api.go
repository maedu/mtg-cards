package api

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/maedu/mtg-cards/user/auth"
	"github.com/maedu/mtg-cards/user/db"
	"github.com/maedu/mtg-cards/util"
)

var usernamePattern, _ = regexp.Compile("^[a-zA-Z0-9-]+$")

// Setup Setup REST API
func Setup(r *gin.Engine) {
	r.GET("/api/user", handleGetUser)
	r.POST("/api/user/rename/:newName", handleRenameUser)
	r.GET("/api/user/cards", handleGetCards)
}

func handleGetUser(c *gin.Context) {
	if userID, ok := auth.GetUserIDFromAccessToken(c, true); ok {
		collection, err := db.GetUserCollection()
		defer collection.Disconnect()
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}

		user, err := collection.GetUserByUserID(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}

		if user == nil {
			userName, err := calcUserName(userID, &collection)
			if err != nil {
				c.JSON(http.StatusInternalServerError, err)
				return
			}
			user = &db.User{
				UserID:   userID,
				UserName: userName,
			}
			_, err = collection.Create(user)
			if err != nil {
				c.JSON(http.StatusInternalServerError, err)
				return
			}
		}

		c.JSON(http.StatusOK, user)
		return
	}
	c.JSON(http.StatusUnauthorized, nil)
}

func handleRenameUser(c *gin.Context) {
	if userID, ok := auth.GetUserIDFromAccessToken(c, true); ok {

		collection, err := db.GetUserCollection()
		defer collection.Disconnect()
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}

		userName := c.Param("newName")
		if !validUsername(userName) {
			c.JSON(http.StatusBadRequest, "Username is invalid. Min lenght is 5 and only letter, numbers, - and _ allowed.")
			return
		}

		available, err := usernameAvailable(userName, userID, &collection)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		if !available {
			c.JSON(http.StatusBadRequest, "Username already taken")
			return
		}

		user, err := collection.GetUserByUserID(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}

		if user == nil {
			user = &db.User{
				UserID:   userID,
				UserName: userName,
			}
			_, err = collection.Create(user)
			if err != nil {
				c.JSON(http.StatusInternalServerError, err)
				return
			}
		} else {
			user.UserName = userName
			user, err = collection.Update(user)
			if err != nil {
				c.JSON(http.StatusInternalServerError, err)
				return
			}
		}

		c.JSON(http.StatusOK, user)
		return
	}
	c.JSON(http.StatusUnauthorized, nil)
}

func calcUserName(userID string, collection *db.UserCollection) (string, error) {
	var userName string
	for {
		userName = fmt.Sprintf("user-%s", util.RandomString(5))
		available, err := usernameAvailable(userName, userID, collection)
		if err != nil {
			return "", err
		}
		if available {
			break
		}
	}
	return userName, nil
}

func validUsername(userName string) bool {
	if len(userName) < 5 {
		return false
	}
	return usernamePattern.MatchString(userName)
}

func usernameAvailable(userName string, userID string, collection *db.UserCollection) (bool, error) {
	user, err := collection.GetUserByUserName(userName)
	if err != nil {
		return true, err
	}
	if user == nil {
		return true, nil
	}
	if user.UserID == userID {
		return true, nil
	}

	return false, nil
}

func handleGetCards(c *gin.Context) {
	if userID, ok := auth.GetUserIDFromAccessToken(c, true); ok {
		collection, err := db.GetUserCardCollection()
		defer collection.Disconnect()
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		userCards, err := collection.GetUserCardsByUserID(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, userCards)
		return
	}
}
