package auth

import "github.com/gin-gonic/gin"

type UserContext struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func LoginService(context *gin.Context) {
	//获取用户提交登录账号和密码
	var user UserContext
	if err := context.ShouldBindJSON(&user); err != nil {
		context.JSON(401, gin.H{
			"code":  401,
			"error": err.Error(),
		})
		return
	}
	//验证用户名和密码是否正确
	s, err := login(user.Username, user.Password)
	if err != nil {
		context.JSON(401, gin.H{
			"code":  401,
			"error": err.Error(),
		})
		return
	}
	context.JSON(200, gin.H{"aBgiToken": s})
}
