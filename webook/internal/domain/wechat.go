package domain

// WechatInfo 表示用户的微信信息
type WechatInfo struct {
	UnionId string // 微信开放平台的唯一标识,用于关联同一个微信用户在不同应用中的账号
	OpenId  string // 微信公众平台的唯一标识,用于标识用户在特定应用中的身份
}
