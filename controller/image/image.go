package image

import (
	"GopherAI/common/code"
	"GopherAI/controller"
	"GopherAI/service/image"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type (
	RecognizeImageResponse struct {
		ClassName string `json:"class_name,omitempty"` // 识别结果
		controller.Response
	}
)

func RecognizeImage(c *gin.Context) {
	res := new(RecognizeImageResponse)
	file, err := c.FormFile("image")
	if err != nil {
		log.Println("FormFile fail ", err)
		c.JSON(http.StatusOK, res.CodeOf(code.CodeInvalidParams))
		return
	}

	className, err := image.RecognizeImage(file)
	if err != nil {
		log.Println("RecognizeImage fail ", err)
		if image.IsModelNotFoundError(err) {
			// 模型文件不存在，返回更具体的错误信息
			res.CodeOf(code.CodeServerBusy)
			res.StatusMsg = "模型文件未配置，图片识别功能暂不可用"
			c.JSON(http.StatusOK, res)
			return
		}
		c.JSON(http.StatusOK, res.CodeOf(code.CodeServerBusy))
		return
	}

	res.Success()
	res.ClassName = className
	c.JSON(http.StatusOK, res)
}
