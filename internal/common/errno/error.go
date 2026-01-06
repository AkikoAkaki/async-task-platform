// Package errno 提供了系统统一的业务错误定义与处理逻辑。
// 该包设计用于标准化微服务间的错误传递，支持 gRPC 状态码透传及 HTTP 响应转换。
package errno

import "fmt"

// Error 定义了业务逻辑错误的通用结构。
// @Description 包含业务层面的错误码与人类可读的错误信息。
type Error struct {
	Code    int32  `json:"code"`    // 业务自定义错误码
	Message string `json:"message"` // 对外展示的错误描述
}

// Error 实现了标准库的 error 接口，返回格式化的错误字符串。
func (e *Error) Error() string {
	return fmt.Sprintf("code: %d, message: %s", e.Code, e.Message)
}

// New 返回一个初始化的 *Error 实例。
// 建议在定义业务模块特定的错误常量时调用此函数。
func New(code int32, msg string) *Error {
	return &Error{
		Code:    code,
		Message: msg,
	}
}

// 通用错误定义。
// 错误码分配原则：
//   - 0: 成功
//   - 1xx: 系统级别错误 (如数据库、网络异常)
//   - 2xx: 业务逻辑错误 (如资源不存在、权限不足)
var (
	// 0：OK 表示请求成功处理。
	OK = New(0, "success")

	// 10001：服务器内部未知异常，通常对应 HTTP 500。
	ErrInternalServerError = New(10001, "internal server error")
	// 10002：输入参数校验失败。
	ErrInvalidParam = New(10002, "invalid parameter")

	// 20001：请求的任务资源在系统中不存在。
	ErrTaskNotFound = New(20001, "task not found")
	// 20002：尝试创建已存在的任务资源。
	ErrTaskAlreadyExist = New(20002, "task already exists")
)
