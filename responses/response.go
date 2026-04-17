package responses

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应格式
type Response struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	TraceID   string      `json:"traceId,omitempty"`
	Timestamp int64       `json:"timestamp,omitempty"`
}

// PageData 分页数据
type PageData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// Success 成功响应
func Success(data interface{}) Response {
	return Response{
		Code:    0,
		Message: "success",
		Data:    data,
	}
}

// SuccessWithPage 分页成功响应
func SuccessWithPage(list interface{}, total int64, page, pageSize int) Response {
	return Response{
		Code:    0,
		Message: "success",
		Data: PageData{
			List:     list,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
	}
}

// Error 错误响应
func Error(code int, message string) Response {
	return Response{
		Code:    code,
		Message: message,
		Data:    nil,
	}
}

// SendJSON 发送 JSON 响应
func SendJSON(c *gin.Context, statusCode int, resp Response) {
	if traceID := c.GetString("traceID"); traceID != "" {
		resp.TraceID = traceID
	}
	c.JSON(statusCode, resp)
}

// SendSuccess 发送成功响应
func SendSuccess(c *gin.Context, data interface{}) {
	SendJSON(c, http.StatusOK, Success(data))
}

// SendCreated 发送创建成功响应
func SendCreated(c *gin.Context, data interface{}) {
	SendJSON(c, http.StatusCreated, Success(data))
}

// SendNotFound 发送未找到响应
func SendNotFound(c *gin.Context, message string) {
	SendJSON(c, http.StatusNotFound, Error(404, message))
}

// SendBadRequest 发送错误请求响应
func SendBadRequest(c *gin.Context, message string) {
	SendJSON(c, http.StatusBadRequest, Error(400, message))
}

// SendInternalError 发送内部错误响应
func SendInternalError(c *gin.Context, message string) {
	SendJSON(c, http.StatusInternalServerError, Error(500, message))
}

// SendConflict 发送冲突响应
func SendConflict(c *gin.Context, message string) {
	SendJSON(c, http.StatusConflict, Error(409, message))
}
