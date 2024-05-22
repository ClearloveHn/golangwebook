package domain

import (
	"time"
)

// Article 表示一篇文章的结构体
type Article struct {
	Id      int64         // 文章ID
	Title   string        // 文章标题
	Content string        // 文章内容
	Author  Author        // 文章作者
	Status  ArticleStatus // 文章状态
	Ctime   time.Time     // 文章创建时间
	Utime   time.Time     // 文章更新时间
}

// Abstract 返回文章的摘要
func (a Article) Abstract() string {
	str := []rune(a.Content)
	if len(str) > 128 {
		str = str[:128]
	}
	return string(str)
}

// ArticleStatus 表示文章的状态,使用uint8类型
type ArticleStatus uint8

// ToUint8 将ArticleStatus转换为uint8类型
func (s ArticleStatus) ToUint8() uint8 {
	return uint8(s)
}

const (
	// ArticleStatusUnknown 这是一个未知状态
	ArticleStatusUnknown = iota

	// ArticleStatusUnpublished 未发表
	ArticleStatusUnpublished

	// ArticleStatusPublished 已发表
	ArticleStatusPublished

	// ArticleStatusPrivate 仅自己可见
	ArticleStatusPrivate
)

// Author 表示文章作者的结构体
type Author struct {
	Id   int64  // 作者ID
	Name string // 作者名字
}
