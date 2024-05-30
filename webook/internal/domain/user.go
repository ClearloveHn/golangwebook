package domain

import "time"

// User 表示用户信息
type User struct {
	Id         int64      // 用户ID,用于唯一标识一个用户
	Email      string     // 用户的电子邮件地址
	Password   string     // 用户的密码,通常是经过加密的
	Nickname   string     // 用户的昵称,用于展示给其他用户
	Birthday   time.Time  // 用户的生日,格式可以是字符串类型,如"2006-01-02"
	AboutMe    string     // 用户的自我介绍或简介
	Phone      string     // 用户的电话号码
	Ctime      time.Time  // 用户的创建时间
	Utime      time.Time  // 用户的更新时间
	WechatInfo WechatInfo // 用户的微信信息
}
