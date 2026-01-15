package auth

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

type MyClaims struct {
	Phone string `json:"phone"`
	jwt.RegisteredClaims
}

// 不在包级直接依赖外部变量，提供显式设置函数
var MySecret []byte

// 在程序启动时调用 SetSecret(...) 来设置
func SetSecret(s string) {
	MySecret = []byte(s)
}

// Secret 返回一个 jwt.Keyfunc，并校验签名算法类型以防 alg 攻击
func Secret() jwt.Keyfunc {
	return func(token *jwt.Token) (interface{}, error) {
		// 确保使用的是 HMAC 系列签名方法（你在签发时使用的是 HS256）
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		if len(MySecret) == 0 {
			return nil, errors.New("jwt secret is not set")
		}
		return MySecret, nil
	}
}

// 这里传入的是手机号，因为项目登录用的是手机号和密码
func MakeToken(phone string) (tokenString string, err error) {
	if len(MySecret) == 0 {
		return "", errors.New("jwt secret is not set, call SetSecret first")
	}

	claim := MyClaims{
		Phone: phone,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(360 * time.Hour)), // 过期时间：3小时（按需调整）
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim) // 使用 HS256 算法
	tokenString, err = token.SignedString(MySecret)
	return tokenString, err
}

// 解析 jwt
func ParseToken(tokenStr string) (*MyClaims, error) {
	if tokenStr == "" {
		return nil, errors.New("token is empty")
	}

	token, err := jwt.ParseWithClaims(tokenStr, &MyClaims{}, Secret())
	// 如果解析返回错误，尽量返回友好信息
	if err != nil {
		// 如果是 ValidationError，可以细化原因
		if ve, ok := err.(*jwt.ValidationError); ok {
			switch {
			case ve.Errors&jwt.ValidationErrorMalformed != 0:
				return nil, errors.New("非法令牌")
			case ve.Errors&jwt.ValidationErrorExpired != 0:
				return nil, errors.New("令牌过期")
			case ve.Errors&jwt.ValidationErrorNotValidYet != 0:
				return nil, errors.New("令牌未激活")
			default:
				return nil, errors.New("令牌未知错误")
			}
		}
		// 其它错误直接返回
		return nil, err
	}

	// token 可能为 nil（尽管上面没有错误），要检查
	if token == nil {
		return nil, errors.New("token is nil after parsing")
	}

	// 断言并返回 claims
	if claims, ok := token.Claims.(*MyClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("couldn't handle this token")
}
