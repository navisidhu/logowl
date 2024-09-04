package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/navisidhu/logowl/internal/keys"
	"github.com/navisidhu/logowl/internal/models"
	"github.com/navisidhu/logowl/internal/store"
	"github.com/navisidhu/logowl/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
)

type InterfaceAuth interface {
	CreateJWT(string) (string, int64, error)
	ResetPassword(user models.User) (string, error)
	InvalidatePasswordResetToken(email, token string) (bool, error)
}

type Auth struct {
	Store   store.InterfaceStore
	Request InterfaceRequest
}

func (a *Auth) CreateJWT(id string) (string, int64, error) {
	timestamp := time.Now().Unix()
	expiresAt := timestamp + int64(time.Hour.Seconds()*keys.SESSION_TIMEOUT_IN_HOURS)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  id,
		"iat": timestamp,
		"exp": expiresAt,
	})

	signedToken, err := token.SignedString([]byte(keys.GetKeys().SECRET))
	if err != nil {
		return "", 0, err
	}

	return signedToken, expiresAt * 1000, nil
}

func (a *Auth) ResetPassword(user models.User) (string, error) {
	resetToken, err := utils.GenerateRandomString(50)
	if err != nil {
		return "", errors.New("an error occured while creating a password reset token")
	}

	passwordResetToken := models.PasswordResetToken{
		Email:     user.Email,
		Token:     resetToken,
		Used:      false,
		ExpiresAt: time.Now().Unix() + 60,
		CreatedAt: time.Now(),
	}

	_, err = a.Store.PasswordResetTokens().InsertOne(passwordResetToken)
	if err != nil {
		return "", err
	}

	data := map[string]interface{}{
		"FirstName": user.FirstName,
		"URL":       fmt.Sprintf("%s/auth/newpassword?token=%s", keys.GetKeys().CLIENT_URL, resetToken),
	}

	err = a.Request.SendEmail(user.Email, "resetPassword", data)
	if err != nil {
		return "", err
	}

	return passwordResetToken.Token, nil
}

func (a *Auth) InvalidatePasswordResetToken(email, token string) (bool, error) {
	if email == "" || token == "" {
		return false, errors.New("email or token were not provided")
	}

	_, err := a.Store.PasswordResetTokens().FindOneAndUpdate(
		bson.M{"email": email, "token": token, "used": false},
		bson.M{"$set": bson.M{"used": true}},
	)
	if err != nil {
		return false, err
	}

	return true, nil
}

func GetAuthService(store store.InterfaceStore) Auth {
	return Auth{store, &Request{}}
}
