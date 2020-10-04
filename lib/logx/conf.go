package logx

type LogConf struct {
	ServiceName         string `json:",optional"`                              // 服务名称，可选
	Mode                string `json:",default=console,options=console|file"`  // 日志模式，默认命令行，支持命令行或文件
	Path                string `json:",default=logs"`                          // 存储目录，默认放在 logs 目录
	Level               string `json:",default=info,options=info|error|fatal"` // 日志级别，默认信息级，可选信息级|错误级|重大级
	Compress            bool   `json:",optional"`                              // 是否压缩日志，可选
	KeepDays            int    `json:",optional"`                              // 日志保留天数，可选
	StackCooldownMillis int    `json:",default=100"`                           // 堆栈冷却毫秒数，默认100毫秒
}
