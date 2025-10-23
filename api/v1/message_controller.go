package v1

import (
	"github.com/gin-gonic/gin"
	"kama_chat_server/internal/dto/request"
	"kama_chat_server/internal/service/gorm"
	"kama_chat_server/pkg/constants"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"kama_chat_server/internal/config"
)

// GetMessageList 获取聊天记录
func GetMessageList(c *gin.Context) {
	var req request.GetMessageListRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": constants.SYSTEM_ERROR,
		})
		return
	}
	message, rsp, ret := gorm.MessageService.GetMessageList(req.UserOneId, req.UserTwoId)
	JsonBack(c, message, ret, rsp)
}

// GetGroupMessageList 获取群聊消息记录
func GetGroupMessageList(c *gin.Context) {
	var req request.GetGroupMessageListRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": constants.SYSTEM_ERROR,
		})
		return
	}
	message, rsp, ret := gorm.MessageService.GetGroupMessageList(req.GroupId)
	JsonBack(c, message, ret, rsp)
}

// UploadAvatar 上传头像
func UploadAvatar(c *gin.Context) {
	message, ret := gorm.MessageService.UploadAvatar(c)
	JsonBack(c, message, ret, nil)
}

// UploadFile 上传头像
func UploadFile(c *gin.Context) {
	message, ret := gorm.MessageService.UploadFile(c)
	JsonBack(c, message, ret, nil)
}

// DownloadFile 下载文件
func DownloadFile(c *gin.Context) {
	// 获取文件名参数
	fileName := c.Param("filename")
	if fileName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "文件名不能为空",
		})
		return
	}

	// URL解码文件名，处理中文编码
	decodedFileName, err := url.QueryUnescape(fileName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "文件名解码失败",
		})
		return
	}

	// 构建文件路径
	filePath := filepath.Join(config.GetConfig().StaticFilePath, decodedFileName)
	
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "文件不存在",
		})
		return
	}

	// 设置响应头，确保中文文件名正确显示
	c.Header("Content-Disposition", "attachment; filename*=UTF-8''"+url.QueryEscape(decodedFileName))
	c.Header("Content-Type", "application/octet-stream")
	
	// 发送文件
	c.File(filePath)
}
