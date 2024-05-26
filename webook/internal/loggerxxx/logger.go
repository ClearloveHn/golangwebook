package loggerxxx

import "go.uber.org/zap"

// 日志记录器

// Logger 是一个全局的日志记录器变量
// 它可以在整个应用程序中使用,用于记录一般性的日志信息
var Logger *zap.Logger

// CommonLogger 是一个全局的日志记录器变量
// 它用于记录常规的日志信息,例如程序的运行状态、业务逻辑的执行过程等
var CommonLogger *zap.Logger

// SensitiveLogger 是一个全局的日志记录器变量
// 它用于记录敏感的日志信息,例如用户的个人信息、关键的业务数据等
var SensitiveLogger *zap.Logger
