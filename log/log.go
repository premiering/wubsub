// Cool colors since we're cool
package log

import (
	"fmt"

	"github.com/fatih/color"
)

var infoColor = color.New(color.FgHiGreen, color.Bold, color.BgBlack)
var errorColor = color.New(color.FgRed, color.Bold)
var debugColor = color.New(color.FgMagenta, color.Bold)

func InfoLog(format string) {
	fmt.Println(infoColor.Sprint("wubsub info | ") + color.WhiteString(format))
}

func ErrorLog(format string, a ...interface{}) {
	fmt.Println(errorColor.Sprint("wubsub error | ") + color.WhiteString(format))
}

func DebugLog(format string, a ...interface{}) {
	fmt.Println(debugColor.Sprint("wubsub debug | ") + color.WhiteString(format))
}
