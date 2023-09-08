package auth

const (
	TOKEN_KEY   = "token"
	USER_ID_KEY = "userId"
)

type UserInterface interface {
	GetId() string
}

var _authFunc AuthFunc

type AuthFunc func(token string) (user UserInterface, err error)

func RegisterAuthFunc(authFunc AuthFunc) {
	_authFunc = authFunc
}

func GetAuthFunc() (AuthFunc, bool) {
	if _authFunc == nil {
		return nil, false
	}
	return _authFunc, true
}

func GetAuthKey() string {
	return TOKEN_KEY
}
