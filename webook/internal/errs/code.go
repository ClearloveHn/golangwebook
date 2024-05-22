package errs

// 表示不同模块中的错误码

// User 相关的错误码
const (
	// UserInvalidInput 表示用户模块的输入错误,常量值为 401001
	UserInvalidInput = 401001

	// UserInvalidOrPassword 表示用户名错误或者密码不正确,常量值为 401002
	UserInvalidOrPassword = 401002

	// UserDuplicateEmail 表示用户邮箱冲突,常量值为 401003
	UserDuplicateEmail = 401003

	// UserInternalServerError 表示用户模块的系统内部错误,常量值为 501001
	UserInternalServerError = 501001
)

// Article 相关的错误码
const (
	// ArticleInvalidInput 表示文章模块的输入错误,常量值为 402001
	ArticleInvalidInput = 402001

	// ArticleInternalServerError 表示文章模块的系统内部错误,常量值为 502001
	ArticleInternalServerError = 502001
)
