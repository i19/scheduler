package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"scheduler/pkg/errors"
)

type ResponseProtocol struct {
	Code    errors.ErrorCode `json:"code"`    // business code, 0 means ok
	Message string           `json:"message"` // error message
	Data    interface{}      `json:"data"`    // result
}

func Response(c *gin.Context, code errors.ErrorCode, errorMessage string, data interface{}) {
	c.JSON(http.StatusOK, ResponseProtocol{
		Code:    code,
		Message: errorMessage,
		Data:    data,
	})
}
func ResponseOK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, ResponseProtocol{
		Code:    errors.OK,
		Message: "success",
		Data:    data,
	})
}
