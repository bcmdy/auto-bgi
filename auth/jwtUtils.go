package auth

import (
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

type MyClaims struct {
	Phone                string `json:"phone"`
	jwt.RegisteredClaims        // 注意!这是jwt-go的v4版本新增的，原先是jwt.StandardClaims
}

var MySecret = []byte(User.Auth.Password) // 定义secret，后面会用到

func Secret() jwt.Keyfunc {
	return func(token *jwt.Token) (interface{}, error) {
		return []byte(MySecret), nil // 这是我的secret
	}
}

// 这里传入的是手机号，因为我项目登陆用的是手机号和密码
func MakeToken(phone string) (tokenString string, err error) {
	claim := MyClaims{
		Phone: phone,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(62 * time.Hour * time.Duration(1))), // 过期时间3小时
			IssuedAt:  jwt.NewNumericDate(time.Now()),                                        // 签发时间
			NotBefore: jwt.NewNumericDate(time.Now()),                                        // 生效时间
		}}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim) // 使用HS256算法
	tokenString, err = token.SignedString(MySecret)
	return tokenString, err
}

// 解析jwt
func ParseToken(tokenss string) (*MyClaims, error) {
	token, err := jwt.ParseWithClaims(tokenss, &MyClaims{}, Secret())
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, errors.New("非法令牌")
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, errors.New("令牌国企")
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return nil, errors.New("令牌未激活")
			} else {
				return nil, errors.New("令牌未知错误")
			}
		}
	}
	if claims, ok := token.Claims.(*MyClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("couldn't handle this token")
}
