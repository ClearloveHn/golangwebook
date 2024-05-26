package startup

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// 初始化数据库连接

func InitDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13316)/webook"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	// 初始化数据库表结构
	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}
