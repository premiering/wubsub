// Cool colors since we're cool
package log

import (
	"fmt"
	"time"
)

var green = "\x1B[32m\x1B[1m"
var reset = "\033[0m"
var red = "\x1B[31m\x1B[1m"
var magenta = "\x1B[35m\x1B[1m"
var gray = "\x1B[90m"

func InfoLog(format string, a ...interface{}) {
	log("info", green, format, a)
}

func ErrorLog(format string, a ...interface{}) {
	log("error", red, format, a)
}

func DebugLog(format string, a ...interface{}) {
	log("debug", magenta, format, a)
}

func log(title string, color string, format string, a []interface{}) {
	t := time.Now()
	tstr := t.Format("15:04:05")
	if len(a) > 0 {
		format = fmt.Sprintf(format, a...)
	}
	fmt.Println(color + "wubsub " + title + " | " + gray + tstr + " |" + reset + " " + format)
}
