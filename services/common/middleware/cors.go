package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

/*
默认规则（无 CORS 时）： 如果在浏览器中打开了 https://we-circle-h5.zeabur.app，
那么网页里的 JavaScript 默认只能向 https://we-circle-h5.zeabur.app 这个完全相同的域名发起请求。
但user-service的地址是https://we-circle.zeabur.app，所以需要设置CORS。
*/

// CORS 处理跨域资源共享 (Cross-Origin Resource Sharing) 的中间件
// 浏览器出于安全考虑，默认遵循同源策略，限制从脚本内发起的跨源HTTP请求。
// 此中间件通过设置响应头，指示浏览器允许跨域请求。
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置允许跨域的响应头
		// Access-Control-Allow-Origin: 允许访问的源（域名）。
		// "*" 表示允许任何域名访问。在生产环境中，出于安全考虑，建议设置为具体的域名（如 https://we-circle.zeabur.app）。
		c.Header("Access-Control-Allow-Origin", "*")

		// Access-Control-Allow-Methods: 允许的 HTTP 请求方法。
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")

		// Access-Control-Allow-Headers: 允许在请求中携带的自定义头信息。
		// Content-Type: 请求体格式
		// Authorization: 认证 Token
		// X-CSRF-Token: CSRF 防护
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		// Access-Control-Expose-Headers: 允许浏览器（前端代码）访问的响应头。
		// 默认情况下，浏览器只能访问基本的响应头（如 Cache-Control, Content-Type 等）。
		c.Header("Access-Control-Expose-Headers", "Content-Length")

		// Access-Control-Allow-Credentials: 是否允许请求携带凭证（如 Cookie、HTTP 认证信息）。
		// 如果设置为 "true"，则 Access-Control-Allow-Origin 不能为 "*"，必须指定具体的域名（但在某些配置下 "*" 也可能生效，视浏览器实现而定，规范建议不能为 "*"）。
		// 注意：当前配置为 "*" 且 Allow-Credentials 为 true 时，部分浏览器可能会报错。
		// 实际开发中如果前端是 we-circle.zeabur.app，后端建议动态获取 Origin 并设置。
		c.Header("Access-Control-Allow-Credentials", "true")

		// 处理 OPTIONS 请求（预检请求 Preflight Request）
		// 浏览器在发送复杂请求（如带自定义头、非简单方法）前，会先发送一个 OPTIONS 请求询问服务器是否允许。
		// 此处直接返回 204 No Content，表示允许该请求，无需进一步处理。
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		// 继续处理后续请求
		c.Next()
	}
}
