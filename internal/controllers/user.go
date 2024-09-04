package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/navisidhu/logowl/internal/models"
	"github.com/navisidhu/logowl/internal/services"
	"github.com/navisidhu/logowl/internal/store"
	"github.com/navisidhu/logowl/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserControllers struct {
	UserService services.InterfaceUser
}

func (u *UserControllers) Get(c *gin.Context) {
	userData, ok := c.Get("user")
	if !ok {
		utils.RespondWithError(c, http.StatusInternalServerError, "could not parse user data")
		return
	}

	userDetails, err := u.UserService.FetchAllInformation(bson.M{"_id": userData.(models.User).ID})
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	if !userDetails.IsOrganizationOwner {
		userDetails.Organization.SubscriptionID = ""
	}

	userDetails.Password = ""

	utils.RespondWithJSON(c, userDetails)
}

func (u *UserControllers) Invite(c *gin.Context) {
	userData, ok := c.Get("user")
	if !ok {
		utils.RespondWithError(c, http.StatusInternalServerError, "could not parse user data")
		return
	}

	if userData.(models.User).Role != "admin" {
		utils.RespondWithError(c, http.StatusForbidden, "you need to be admin to invite new users")
		return
	}

	var newUser models.User

	err := json.NewDecoder(c.Request.Body).Decode(&newUser)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	newUser.OrganizationID = userData.(models.User).OrganizationID

	persistedUser, err := u.UserService.Invite(newUser)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(c, persistedUser)
}

func (u *UserControllers) Delete(c *gin.Context) {
	userData, ok := c.Get("user")
	if !ok {
		utils.RespondWithError(c, http.StatusInternalServerError, "could not parse user data")
		return
	}

	if userData.(models.User).Role != "admin" {
		utils.RespondWithError(c, http.StatusForbidden, "you need to be admin to delete users")
		return
	}

	userID := c.Param("id")

	parsedUserID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "the provided user ID is invalid")
		return
	}

	filter := bson.M{"_id": parsedUserID, "organizationId": userData.(models.User).OrganizationID, "isOrganizationOwner": false}

	deleteCount, err := u.UserService.Delete(filter)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	if deleteCount == 0 {
		utils.RespondWithError(c, http.StatusBadRequest, "the user with the ID "+userID+" does not exist or can not be deleted")
		return
	}

	utils.RespondWithSuccess(c)
}

func (u *UserControllers) DeleteUserAccount(c *gin.Context) {
	userData, ok := c.Get("user")
	if !ok {
		utils.RespondWithError(c, http.StatusInternalServerError, "could not parse user data")
		return
	}

	if userData.(models.User).IsOrganizationOwner {
		utils.RespondWithError(c, http.StatusForbidden, "you can not delete your account as organization owner")
		return
	}

	deleteCount, err := u.UserService.Delete(bson.M{"_id": userData.(models.User).ID})
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	if deleteCount == 0 {
		utils.RespondWithError(c, http.StatusBadRequest, "could not delete user")
		return
	}

	utils.RespondWithSuccess(c)
}

func GetUserController(store store.InterfaceStore) UserControllers {
	userService := services.GetUserService(store)

	return UserControllers{
		UserService: &userService,
	}
}
