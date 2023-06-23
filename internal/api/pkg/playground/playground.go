package playground

import (
	"fmt"
	"runtime"

	"github.com/matishsiao/goInfo"
)

// Playground for testing stuff.
func Playground() string {
	html := "<html><head><title>Go Playground</title></head><body><h1>Go</h1>"

	info, _ := goInfo.GetInfo()

	html = html + "GoOS:" + info.GoOS + "<br>"
	html = html + "Kernel:" + info.Kernel + "<br>"
	html = html + "Core:" + info.Core + "<br>"
	html = html + "Platform:" + info.Platform + "<br>"
	html = html + "OS:" + info.OS + "<br>"
	html = html + "Hostname:" + info.Hostname + "<br>"
	html = html + "CPUs:" + fmt.Sprint(info.CPUs) + "<br>"
	html = html + "Arch:" + runtime.GOARCH + "<br>"

	html = html + "</body></html>"
	return html
}
