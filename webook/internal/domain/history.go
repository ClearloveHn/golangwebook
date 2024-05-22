package domain

// HistoryRecord 表示一条历史记录
type HistoryRecord struct {
	BizId int64  // 业务ID,用于标识不同的业务类型
	Biz   string // 业务名称,用于描述业务的具体内容
	Uid   int64  // 用户ID,表示该历史记录所属的用户
}
