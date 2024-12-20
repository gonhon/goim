package golog

import (
	"flag"

	"github.com/golang/glog"
)

// v=0: 这是默认级别，表示只记录错误和重要信息。
// v=1: 记录一些调试信息，适合一般调试用途。
// v=2: 记录更详细的调试信息，可以用于更细致的排查问题。
// v=3: 记录非常详细的调试信息，通常是开发过程中使用。
// v=4: 记录极其详细的信息，适合深入分析和调试。

// log_dir 和 logtostderr 只能存在一个
// 且log_dir目录不存在时，不会创建文件夹

func logs() {
	// flag.Bool("logtostderr", true, "log to standard error instead of files")
	flag.Set("log_dir", "./logs") // 输出到控制台
	flag.Set("v", "2")            // 设置日志级别
	flag.Parse()

	glog.Info("This is an info message.")
	glog.Warning("This is a warning message.")
	glog.Error("This is an error message.")
	if glog.V(2) {
		glog.Info("Starting transaction...")
	}

	defer glog.Flush()
}
