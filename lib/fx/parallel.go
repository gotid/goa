package fx

import "github.com/z-sdk/goa/lib/threading"

// Parallel 并行且安全的执行一组函数
func Parallel(fns ...func()) {
	group := threading.NewRoutineGroup()
	for _, fn := range fns {
		group.RunSafe(fn)
	}
	group.Wait()
}
