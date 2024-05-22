package domain

// AsyncSMS 表示一个异步发送的短信任务
type AsyncSMS struct {
	Id       int64    // 短信任务的唯一标识
	TplId    int64    // 短信模板的ID
	Args     []string // 短信模板的参数列表
	Numbers  []string // 短信接收者的手机号列表
	RetryMax int      // 最大重试次数,如果发送失败,会自动重试直到达到最大次数
}
