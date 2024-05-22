package domain

// Interactive 表示一个交互数据的结构体
type Interactive struct {
	BizId      int64 // 业务ID,用于标识不同的业务类型,如文章、评论等
	ReadCnt    int64 // 阅读数,表示该业务被阅读的次数
	LikeCnt    int64 // 点赞数,表示该业务被点赞的次数
	CollectCnt int64 // 收藏数,表示该业务被收藏的次数
	Liked      bool  // 是否已点赞,表示当前用户是否对该业务点过赞
	Collected  bool  // 是否已收藏,表示当前用户是否已收藏该业务
}
