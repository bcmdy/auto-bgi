package CDCollectionManagement

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

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

type CreateFolderRequest struct {
	ParentPath string `json:"parentPath"`
	FolderName string `json:"folderName"`
}

type CreateScriptRequest struct {
	ParentPath string `json:"parentPath"`
	ScriptPath string `json:"scriptPath"`
}

type DeleteNodeRequest struct {
	TargetPath string `json:"targetPath"`
}

func CreateFolder(context *gin.Context) {
	var req CreateFolderRequest
	if err := context.ShouldBindJSON(&req); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "参数格式错误", "error": err.Error()})
		return
	}

	targetPath, err := CreatePathingFolder(req.ParentPath, req.FolderName)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"message": "创建成功",
		"path":    targetPath,
	})
}

func CreateScript(context *gin.Context) {
	var req CreateScriptRequest
	if err := context.ShouldBindJSON(&req); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "参数格式错误", "error": err.Error()})
		return
	}

	targetPath, err := AddPathingScript(req.ParentPath, req.ScriptPath)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"message": "新增成功",
		"path":    targetPath,
	})
}

func DeleteNode(context *gin.Context) {
	var req DeleteNodeRequest
	if err := context.ShouldBindJSON(&req); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "参数格式错误", "error": err.Error()})
		return
	}

	targetPath, err := DeletePathingNode(req.TargetPath)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"message": "删除成功",
		"path":    targetPath,
	})
}
