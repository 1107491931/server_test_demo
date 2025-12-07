package errs

var MsgFlags = map[int]string{
	SUCCESS:           "success",
	ERROR:             "fail",
	INVALID_PARAMS:    "请求参数错误",
	UNAUTHORIZED:      "未授权",
	NOT_FOUND:         "资源未找到",
	SERVER_ERROR:      "服务器内部错误",
	TOKEN_ERROR:       "Token错误",
	TOKEN_TIMEOUT:     "Token已过期",
	TOKEN_REVOKED:     "Token已失效",
	TOO_MANY_REQUESTS: "请求过多",

	USER_EXISTS:        "用户已存在",
	USER_NOT_FOUND:     "用户不存在",
	PASSWORD_INCORRECT: "密码错误",
	USER_CREATE_FAIL:   "创建用户失败",
	USER_UPDATE_FAIL:   "更新用户失败",

	POST_NOT_FOUND: "动态未找到",
}

// GetMsg 获取错误码对应的消息
func GetMsg(code int) string {
	msg, ok := MsgFlags[code]
	if ok {
		return msg
	}
	return MsgFlags[ERROR]
}
