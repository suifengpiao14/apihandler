package auth

import (
	"github.com/pkg/errors"

	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
)

func CasDoorAuthFunc(jwt string) (user UserInterface, err error) {
	if jwt == "" {
		err = errors.New("token is empty")
		return user, err
	}
	claims, err := casdoorsdk.ParseJwtToken(jwt)
	if err != nil {
		return user, err
	}
	claims.AccessToken = jwt
	return claims.User, nil
}
