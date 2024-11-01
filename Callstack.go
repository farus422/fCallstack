package fCallstack

import (
	"fmt"
	"path"
	"runtime"
	"strings"
	"sync"
)

type SCaller struct {
	Package  string
	Function string
	File     string // full path
	Line     int
}

type SCallstack struct {
	callers []SCaller
}

// 取得物件所獲取的呼叫堆疊資訊
func (cs *SCallstack) GetCallers() []SCaller {
	if (cs.callers == nil) || (len(cs.callers) <= 0) {
		return nil
	}
	return cs.callers
}

// 用Printf()打印出呼叫堆疊
func (cs *SCallstack) Print() {
	for _, caller := range cs.callers {
		fmt.Printf("%s:%d %s()\n", caller.File, caller.Line, caller.Function)
	}
}

// 獲取目前的呼叫堆疊資訊，並且去除掉golang框架的堆疊部分，如果這是個panic，則會從發生panic的地方開始列出
// frontSkip:				從叫用 GetCallstack() 的地方開始，要往上略過多少層，0:叫用GetCallstack()的地方開始列出
// hideTheCallStartFunc:	要隱藏的最上層呼叫者，使之從它以下才會開始出現在呼叫堆疊
func (cs *SCallstack) GetCallstack(frontSkip int, hideTheCallStartFunc string) {
	begin := frontSkip + 2
	callerIndex := 0
	panicFound, panicSearching := false, false
	var n int
	var pcs []uintptr
	var frame runtime.Frame
	var funcname string
	var funcs []string
	for size := 32; ; size += 16 {
		pcs = make([]uintptr, size)
		n = runtime.Callers(0, pcs)
		if n < size {
			break
		}
	}

	frames := runtime.CallersFrames(pcs[:n])
	more := n > 0
	n = 0
	for more {
		frame, more = frames.Next()
		if n > 0 {
			if (hideTheCallStartFunc != "") && (strings.LastIndex(frame.Function, hideTheCallStartFunc) != -1) {
				if callerIndex <= begin {
					callerIndex = n
				}
			} else if (hideTheCallStartFunc == "") && IsDefaultHiddenCaller(frame.Function) {
				if callerIndex <= begin {
					callerIndex = n
				}
			} else if frame.Function == "runtime.goexit" || frame.Function == "testing.tRunner" {
				break
			}
		}
		// 若是系統自動引發panic則會在發生錯誤的地方呼叫panic()，所以必須跳過堆疊上方自動呼叫的部分
		if !panicFound {
			if panicSearching {
				if !strings.HasPrefix(frame.Function, "runtime.") {
					begin = n // 這裡假設叫用者不知道這是panic，所以begin就不加上frontSkip
					panicFound = true
				}
			} else {
				switch frame.Function {
				case "runtime.gopanic", "runtime.panic", "runtime.sigpanic":
					panicSearching = true
				}
			}
		}
		n++
		if frame.Function == "main.main" {
			break
		}
	}
	if begin > n {
		begin = n
	}
	if callerIndex > 0 && callerIndex > begin {
		n = callerIndex
	}

	frames = runtime.CallersFrames(pcs[begin:n])
	n -= begin
	cs.callers = make([]SCaller, n)
	callers := cs.callers
	more = n > 0
	for index := 0; more; index++ {
		frame, more = frames.Next()
		_, funcname = path.Split(frame.Function)
		funcs = strings.SplitN(funcname, ".", 2)
		callers[index].Package = funcs[0]
		callers[index].Function = funcs[len(funcs)-1]
		callers[index].File = frame.File
		callers[index].Line = frame.Line
	}
}

// 獲取目前的呼叫堆疊資訊，並且去除掉golang框架的堆疊部分，配合recover()使用
// frontSkip:				從發生 panic  的地方開始，要往上略過多少層，0:從發生 panic 的地方開始列出
// hideTheCallStartFunc:	要隱藏的最上層呼叫者，使之從它以下才會開始出現在呼叫堆疊
func (cs *SCallstack) GetCallstackWithPanic(frontSkip int, hideTheCallStartFunc string) {
	size := 32
	panicFound, panicSearching := false, false
	begin := frontSkip // 確實不同於GetCallstack
	callerIndex := 0
	var n int
	var pcs []uintptr
	var frame runtime.Frame
	var funcname string
	var funcs []string
	for size > 0 {
		pcs = make([]uintptr, size)
		n = runtime.Callers(0, pcs)
		if n < size {
			frames := runtime.CallersFrames(pcs[:n])
			more := n > 0
			n = 0
			for more {
				frame, more = frames.Next()
				if n > 0 {
					if (hideTheCallStartFunc != "") && (strings.LastIndex(frame.Function, hideTheCallStartFunc) != -1) {
						if callerIndex <= begin {
							callerIndex = n
						}
					} else if (hideTheCallStartFunc == "") && IsDefaultHiddenCaller(frame.Function) {
						if callerIndex <= begin {
							callerIndex = n
						}
					} else if frame.Function == "runtime.goexit" || frame.Function == "testing.tRunner" {
						break
					}
				}
				// 若是系統自動引發panic則會在發生錯誤的地方呼叫panic()，所以必須跳過堆疊上方自動呼叫的部分
				if !panicFound {
					if panicSearching {
						if !strings.HasPrefix(frame.Function, "runtime.") {
							begin = n + frontSkip // 叫用者明確知道這是panic，所以要跳過多少層是由叫用者自己決定
							panicFound = true
						}
					} else {
						switch frame.Function {
						case "runtime.gopanic", "runtime.panic", "runtime.sigpanic":
							panicSearching = true
						}
					}
				}
				n++
				if frame.Function == "main.main" {
					break
				}
			}
			if begin > n {
				begin = n
			}
			if callerIndex > 0 && callerIndex > begin {
				n = callerIndex
			}

			frames = runtime.CallersFrames(pcs[begin:n])
			n -= begin
			cs.callers = make([]SCaller, n)
			callers := cs.callers
			more = n > 0
			for index := 0; more; index++ {
				frame, more = frames.Next()
				_, funcname = path.Split(frame.Function)
				funcs = strings.SplitN(funcname, ".", 2)
				callers[index].Package = funcs[0]
				callers[index].Function = funcs[len(funcs)-1]
				callers[index].File = frame.File
				callers[index].Line = frame.Line
			}
			break
		} else {
			size += 16
		}
	}
}

// 獲取指定堆疊的呼叫函式名稱
func (cs *SCallstack) GetFunctionName(index int) string {
	if len(cs.callers) <= index {
		return ""
	}
	return cs.callers[index].Function
}

// 釋放內部空間以便重用物件
func (cs *SCallstack) Clean() {
	if cs.callers != nil {
		cs.callers = cs.callers[:0]
	}
}

// 獲取目前的呼叫堆疊資訊，並且去除掉golang框架的堆疊部分，如果這是個panic，則會從發生panic的地方開始列出
// frontSkip:				從叫用 GetCallstack() 的地方開始，要往上略過多少層，0:叫用GetCallstack()的地方也會出現在呼叫堆疊中
// hideTheCallStartFunc:	要隱藏的最上層呼叫者，使之從它以下才會開始出現在呼叫堆疊
// 如果您講求效率，那麼您可以自己建立SCallstack並呼叫SCallstack.GetCallstack(frontSkip, hideTheCallStartFunc)
func GetCallstack(frontSkip int, hideTheCallStartFunc string) *SCallstack {
	cs := &SCallstack{}
	cs.GetCallstack(frontSkip+1, hideTheCallStartFunc)
	return cs
}

// 獲取目前的呼叫堆疊資訊，並且去除掉golang框架的堆疊部分，配合recover()使用
// frontSkip:				從發生 panic 的地方開始，要往上略過多少層，0:從發生 panic 的地方開始列出
// hideTheCallStartFunc:	要隱藏的最上層呼叫者，使之從它以下才會開始出現在呼叫堆疊
// 如果您講求效率，那麼您可以自己建立SCallstack並呼叫SCallstack.GetCallstackWithPanic(frontSkip, hideTheCallStartFunc)
func GetCallstackWithPanic(frontSkip int, hideTheCallStartFunc string) *SCallstack {
	cs := &SCallstack{}
	cs.GetCallstackWithPanic(frontSkip, hideTheCallStartFunc)
	return cs
}

type sHiddenFunctions struct {
	mutex     sync.RWMutex
	num       int
	functions []string
}

var hiddenFunctions = sHiddenFunctions{functions: make([]string, 0, 8)}

func AddDefaultHiddenCaller(funcName string) {
	hiddenFunctions.mutex.Lock()
	hiddenFunctions.num++
	hiddenFunctions.functions = append(hiddenFunctions.functions, funcName)
	hiddenFunctions.mutex.Unlock()
}

func IsDefaultHiddenCaller(funcName string) bool {
	hiddenFunctions.mutex.RLock()
	for i := hiddenFunctions.num; i > 0; {
		i--
		if strings.LastIndex(funcName, hiddenFunctions.functions[i]) != -1 {
			hiddenFunctions.mutex.RUnlock()
			return true
		}
	}
	hiddenFunctions.mutex.RUnlock()
	return false
}
