package operator

import (
	"strings"
	"io/ioutil"
	"log"
	"os"
	"github.com/gookit/color"
)

var green = color.FgGreen.Render
var red = color.FgRed.Render


type Log struct {
	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
}

func NewLog(debugLevel string) Log {
	dischargeHandle := ioutil.Discard
	traceHandle := os.Stdout
	infoHandle := os.Stdout
	warningHandle := os.Stdout
	errorHandle := os.Stderr

	l := Log{}

	if strings.ToUpper(debugLevel) == "TRACE"{
	l.Trace = log.New(traceHandle,
		"TRACE: ",
		log.Ldate|log.Ltime)
	} else {
	l.Trace = log.New(dischargeHandle,
		"TRACE: ",
		log.Ldate|log.Ltime)		
	}

	l.Info = log.New(infoHandle,
		green("INFO: "),
		log.Ldate|log.Ltime)

	l.Warning = log.New(warningHandle,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	l.Error = log.New(errorHandle,
		red("ERROR: "),
		log.Ldate|log.Ltime|log.Lshortfile)

	return l
}
