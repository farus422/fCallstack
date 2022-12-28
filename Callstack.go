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

func (cs *SCallstack) GetCallers() []SCaller {
	if (cs.callers == nil) || (len(cs.callers) <= 0) {
		return nil
	}
	return cs.callers
}
func (cs *SCallstack) Print() {
	for _, caller := range cs.callers {
		fmt.Printf("%s:%d %s()\n", caller.File, caller.Line, caller.Function)
	}
}

// frontSkip:				從叫用 GetCallstack() 的地方開始，要往上略過多少層，0:叫用GetCallstack()的地方也會出現在呼叫堆疊中
// hideTheCallStartFunc:	要隱藏的最上層呼叫者，使之從它以下才會開始出現在呼叫堆疊
func (cs *SCallstack) GetCallstack(frontSkip int, hideTheCallStartFunc string) {
	size := 16
	callerIndex := 0
	var n int
	var pcs []uintptr
	var frame runtime.Frame
	var funcname string
	var funcs []string
	for size > 0 {
		pcs = make([]uintptr, size)
		n = runtime.Callers(frontSkip+2, pcs)
		if n < size {
			frames := runtime.CallersFrames(pcs[:n])
			more := n > 0
			n = 0
			for more {
				frame, more = frames.Next()
				if n > 0 {
					if (hideTheCallStartFunc != "") && (strings.LastIndex(frame.Function, hideTheCallStartFunc) != -1) {
						callerIndex = n
					} else if (hideTheCallStartFunc == "") && IsDefaultHiddenCaller(frame.Function) {
						callerIndex = n
					} else if frame.Function == "runtime.goexit" {
						// } else if strings.Compare(frame.Function, "runtime.goexit") == 0 {
						break
					}
				}
				n++
				// if strings.Compare(frame.Function, "main.main") == 0 {
				if frame.Function == "main.main" {
					break
				}
			}
			if callerIndex > 0 {
				n = callerIndex
			}

			cs.callers = make([]SCaller, n)
			frames = runtime.CallersFrames(pcs[:n])
			index := 0
			more = n > 0
			for more {
				frame, more = frames.Next()
				_, funcname = path.Split(frame.Function)
				funcs = strings.SplitN(funcname, ".", 2)
				cs.callers[index].Package = funcs[0]
				cs.callers[index].Function = funcs[1]
				cs.callers[index].File = frame.File
				cs.callers[index].Line = frame.Line
				index++
			}
			break
		} else {
			size += 16
		}
	}
}

// frontSkip:				從叫用 GetCallstack() 的地方開始，要往上略過多少層，0:叫用GetCallstack()的地方也會出現在呼叫堆疊中
// hideTheCallStartFunc:	要隱藏的最上層呼叫者，使之從它以下才會開始出現在呼叫堆疊
func (cs *SCallstack) GetCallstackWithPanic(frontSkip int, hideTheCallStartFunc string) {
	size := 16
	searching := false
	searchdone := false
	begin := 0
	callerIndex := 0
	var n int
	var pcs []uintptr
	var frame runtime.Frame
	var funcname string
	var funcs []string
	for size > 0 {
		pcs = make([]uintptr, size)
		n = runtime.Callers(frontSkip+2, pcs)
		if n < size {
			frames := runtime.CallersFrames(pcs[:n])
			more := n > 0
			n = 0
			for more {
				frame, more = frames.Next()
				if n > 0 {
					if (hideTheCallStartFunc != "") && (strings.LastIndex(frame.Function, hideTheCallStartFunc) != -1) {
						callerIndex = n
					} else if (hideTheCallStartFunc == "") && IsDefaultHiddenCaller(frame.Function) {
						callerIndex = n
						// } else if strings.Compare(frame.Function, "runtime.goexit") == 0 {
					} else if frame.Function == "runtime.goexit" {
						break
					}
				}
				// 若是系統自動引發panic則會在發生錯誤的地方呼叫panic()，所以必須跳過堆疊上方自動呼叫的部分
				if !searchdone {
					if searching {
						if !strings.HasPrefix(frame.Function, "runtime.") {
							begin = n
							searchdone = true
						}
						// } else if (strings.Compare(frame.Function, "runtime.gopanic") == 0) || (strings.Compare(frame.Function, "runtime.panic") == 0) || (strings.Compare(frame.Function, "runtime.sigpanic") == 0) {
					} else if (frame.Function == "runtime.gopanic") || (frame.Function == "runtime.panic") || (frame.Function == "runtime.sigpanic") {
						searching = true
					}
				}
				n++
				// if strings.Compare(frame.Function, "main.main") == 0 {
				if frame.Function == "main.main" {
					break
				}
			}
			if callerIndex > 0 {
				n = callerIndex
			}

			cs.callers = make([]SCaller, n-begin)
			frames = runtime.CallersFrames(pcs[begin:n])
			n -= begin
			index := 0
			more = n > 0
			for more {
				frame, more = frames.Next()
				// cs.callers[index].Function = frame.Function
				_, funcname = path.Split(frame.Function)
				funcs = strings.SplitN(funcname, ".", 2)
				cs.callers[index].Package = funcs[0]
				cs.callers[index].Function = funcs[1]
				cs.callers[index].File = frame.File
				cs.callers[index].Line = frame.Line
				index++
			}
			break
		} else {
			size += 16
		}
	}
}

func (cs *SCallstack) GetFunctionName(index int) string {
	if len(cs.callers) <= index {
		return ""
	}
	return cs.callers[index].Function
}

func (cs *SCallstack) Clean() {
	if cs.callers == nil {
		cs.callers = cs.callers[:0]
	}
}

// frontSkip:				從叫用 GetCallstack() 的地方開始，要往上略過多少層，0:叫用GetCallstack()的地方也會出現在呼叫堆疊中
// hideTheCallStartFunc:	要隱藏的最上層呼叫者，使之從它以下才會開始出現在呼叫堆疊
// 如果您講求效率，那麼您可以自己建立SCallstack並呼叫SCallstack.GetCallstack(frontSkip, hideTheCallStartFunc)
func GetCallstack(frontSkip int, hideTheCallStartFunc string) *SCallstack {
	cs := &SCallstack{}
	cs.GetCallstack(frontSkip+1, hideTheCallStartFunc)
	return cs
}

// frontSkip:				從叫用 GetCallstack() 的地方開始，要往上略過多少層，0:叫用GetCallstack()的地方也會出現在呼叫堆疊中
// hideTheCallStartFunc:	要隱藏的最上層呼叫者，使之從它以下才會開始出現在呼叫堆疊
// 如果您講求效率，那麼您可以自己建立SCallstack並呼叫SCallstack.GetCallstackWithPanic(frontSkip, hideTheCallStartFunc)
func GetCallstackWithPanic(frontSkip int, hideTheCallStartFunc string) *SCallstack {
	cs := &SCallstack{}
	cs.GetCallstackWithPanic(frontSkip+1, hideTheCallStartFunc)
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
