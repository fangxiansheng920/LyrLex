// Package resp 统一响应结构。
package resp

// Body 统一响应体：{code, message, data}。
type Body struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// OK 成功响应，code 固定为 0。
func OK(data interface{}) Body {
	return Body{Code: 0, Message: "ok", Data: data}
}

// Error 失败响应。
func Error(code int, message string) Body {
	return Body{Code: code, Message: message}
}
