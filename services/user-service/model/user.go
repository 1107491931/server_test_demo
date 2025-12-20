package model

import (
	"gorm.io/gorm"
)

// User 用户模型
// 包含用户的基本信息：用户名、邮箱、密码、头像
// json表示User在序列化与反序列化时json中的字段名称
/*
json:"-" 表示： 这个字段在 JSON 序列化（编码）和反序列化（解码）时完全忽略。
// 序列化
user := User{
    Username: "john",
    Email:    "john@example.com",
    Password: "secret123",
}
data, _ := json.Marshal(user)
fmt.Println(string(data)) // 输出：{"username":"john","email":"john@example.com"}, 密码被忽略

// 反序列化
jsonStr := `{"username":"john","email":"john@example.com","password":"hacked"}`
var user User
json.Unmarshal([]byte(jsonStr), &user)
fmt.Printf("Password: %s\n", user.Password)  // "" （空字符串，不会被覆盖）
*/
type User struct {
	gorm.Model
	Username string `gorm:"type:varchar(100);not null" json:"username"`          // 用户名
	Email    string `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"` // 邮箱，唯一索引， 如果插入相同邮箱的user数据时会报错
	Password string `gorm:"type:varchar(255);not null" json:"-"`                 // 密码，json序列化时忽略
	Avatar   string `gorm:"type:varchar(500);default:''" json:"avatar"`          // 头像URL
}

// AutoMigrate 自动迁移数据库表结构
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&User{})
}

/*
数据库的相关类型：
type:varchar(255)      // 字符串，长度255，指的是字符个数，而不是字节长度
type:text              // 长文本
type:integer           // 整数
type:bigint            // 大整数
type:boolean           // 布尔值
type:decimal(10,2)     // 小数，10位总长度，2位小数
type:datetime          // 日期时间

约束
not null               // 非空约束
default:'value'        // 默认值
autoIncrement          // 自增
primaryKey             // 主键
unique                 // 唯一约束
uniqueIndex            // 唯一索引
index                  // 普通索引
*/
