package fCallstack

import (
	"fmt"
	"io"
	"path"
	"runtime"
	"strconv"
	"strings"
)

type SCaller struct {
	Package  string
	Function string
	File     string // full path
	Line     int64
}

func (c *SCaller) FromFrame(frame *runtime.Frame) {
	_, funcname := path.Split(frame.Function)
	names := strings.SplitN(funcname, ".", 2)
	c.Package = names[0]
	c.Function = names[len(names)-1]
	c.File = frame.File
	c.Line = int64(frame.Line)
}

// Format formats the frame according to the fmt.Formatter interface.
//
//	%s    source file
//	%d    source line
//	%n    function name
//	%v    equivalent to %s:%d
//
// Format accepts flags that alter the printing of some verbs, as follows:
//
//	%+s   function name and path of source file relative to the compile time
//	      GOPATH separated by \n\t (<funcname>\n\t<path>)
//	%+v   equivalent to %+s:%d
func (c SCaller) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		switch {
		case s.Flag('+'):
			io.WriteString(s, c.Package)
			io.WriteString(s, ".")
			io.WriteString(s, c.Function)
			io.WriteString(s, "\n\t")
			io.WriteString(s, c.File)
		default:
			io.WriteString(s, path.Base(c.File)) // %v也會叫用這裡，考慮後決定不要顯示完整路徑 [/aa/bb/cc.go:xx /ii/jj/kk.go:xx] 感覺太長了
		}
	case 'd':
		io.WriteString(s, strconv.FormatInt(c.Line, 10))
	case 'n':
		io.WriteString(s, c.Package)
		io.WriteString(s, ".")
		io.WriteString(s, c.Function)
	case 'v':
		c.Format(s, 's')
		io.WriteString(s, ":")
		c.Format(s, 'd')
	}
}
