# 歡迎使用fCallstack套件

### 目錄<span id="目錄"></span>

<a href="#example">fCallback 使用</a><br />
<a href="#object">物件方法</a><br />

---------------------------------------------------------

## fCallback 使用<span id="example"></span>&nbsp;&nbsp;&nbsp;&nbsp;<a href="#目錄">(回到目錄)</a>
範例：列出呼叫堆疊
```golang
package main

import (
	"fmt"

	fcb "github.com/farus422/fCallstack"
)

func FuncB() {
	cs := fcb.GetCallstack(0, "")
	for _, caller := range cs.GetCallers() {
		fmt.Printf("%s:%d %s()\n", caller.File, caller.Line, caller.Function)
	}
}
func FuncA() {
	FuncB()
}
func main() {
	FuncA()
}
```

範例：於panic時列出呼叫堆疊
```golang
package main

import (
	"fmt"

	fcb "github.com/farus422/fCallstack"
)

func FuncB() {
	var p *int
	*p = 100
}
func FuncA() {
	FuncB()
}
func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			cs := fcb.GetCallstackWithPanic(0, "")
			for _, caller := range cs.GetCallers() {
				fmt.Printf("%s:%d %s()\n", caller.File, caller.Line, caller.Function)
			}
			return
		}
	}()
	FuncA()
}
```

## 物件方法<span id="object"></span>&nbsp;&nbsp;&nbsp;&nbsp;<a href="#目錄">(回到目錄)</a>
**獲取目前的呼叫堆疊資訊**
`GetCallstack(frontSkip int, hideTheCallStartFunc string)`
+ frontSkip
 從叫用 GetCallstack() 的地方開始，要往上略過多少層
 給0等於: 叫用GetCallstack()的地方開始列出
+ hideTheCallStartFunc
 要隱藏的最上層呼叫者，使之從它以下才會開始出現在呼叫堆疊

**獲取目前的呼叫堆疊資訊，配合recover()使用**
`GetCallstackWithPanic(frontSkip int, hideTheCallStartFunc string)`
+ frontSkip
 從叫用 GetCallstack() 的地方開始，要往上略過多少層
 給0等於: 叫用GetCallstack()的地方開始列出
+ hideTheCallStartFunc
 要隱藏的最上層呼叫者，使之從它以下才會開始出現在呼叫堆疊

**取得物件所獲取的呼叫堆疊資訊**
`GetCallers() []SCaller`
返回 []SCaller

**用Printf()打印出呼叫堆疊**
`Print()`

**獲取指定堆疊的呼叫函式名稱**
`GetFunctionName(index int) string`
+ index
給定堆疊的index

**釋放內部空間以便重用物件**
`Clean()`
