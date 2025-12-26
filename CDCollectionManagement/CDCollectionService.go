package CDCollectionManagement

import "github.com/gin-gonic/gin"

func CDCollectionRead(context *gin.Context) {
	name := context.Query("name")
	read := ReadRecord(name)
	context.JSON(200, read)
}

func AllUserFile(context *gin.Context) {
	read := ReadAllUser()
	context.JSON(200, read)
}

func ReadPickup(context *gin.Context) {
	name := context.Query("name")
	read := ReadPickupRecord(name)
	context.JSON(200, read)

}
