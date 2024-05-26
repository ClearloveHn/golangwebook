package startup

func InitLogger() logger.LoggerV1 {
	return logger.NewNopLogger()
}
