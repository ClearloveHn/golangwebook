package job

// Job 是一个接口，表示一个可执行的任务
type Job interface {
	// Name 方法返回任务的名称
	// 任务的名称通常用于标识和区分不同的任务
	Name() string

	// Run 方法执行任务的具体逻辑
	// 任务的执行逻辑由具体的实现类来定义
	Run() error
}
