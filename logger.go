package golive

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/logrusorgru/aurora/v3"
)

const (
	LogTrace = iota - 1
	LogDebug
	LogInfo
	LogWarn
	LogError
	LogFatal
	LogPanic
)

type Log func(level int, message string, extra map[string]interface{})

type logEx map[string]interface{}

type LoggerBasic struct {
	Level      int
	Prefix     string
	TimeFormat string
}

func NewLoggerBasic() *LoggerBasic {
	var l LoggerBasic
	l.Level = LogWarn
	l.Prefix = aurora.BrightBlue("GOLIVE ").String()
	l.TimeFormat = aurora.BrightBlack("15:04:05").String()

	return &l
}

func (l *LoggerBasic) Log(level int, message string, extra map[string]interface{}) {
	b := strings.Builder{}
	b.WriteString(l.Prefix)

	b.WriteString(time.Now().Format(l.TimeFormat))

	switch level {
	case LogTrace:
		b.WriteString(aurora.Yellow(" TRACE ").String())
	case LogDebug:
		b.WriteString(aurora.Yellow(" DEBUG ").String())
	case LogInfo:
		b.WriteString(aurora.Red(" WARN  ").String())
	case LogError:
		b.WriteString(aurora.BrightRed(" ERROR ").String())
	case LogFatal:
		b.WriteString(aurora.BrightRed(" FATAL ").String())
	case LogPanic:
		b.WriteString(aurora.BrightRed(" PANIC ").String())
	}

	b.WriteString(message)

	if len(extra) != 0 {
		var keys []string
		for key := range extra {
			keys = append(keys, key)
		}

		sort.Strings(keys)

		for i := 0; i < len(keys); i++ {
			key := keys[i]

			b.WriteString(fmt.Sprintf(" %s%+v", aurora.BrightBlack(key+"=").String(), extra[key]))
		}
	}

	switch level {
	case LogPanic:
		fmt.Println(b.String())
		panic(b.String())
	case LogFatal:
		fmt.Println(b.String())
		os.Exit(1)
	default:
		if level >= l.Level {
			fmt.Println(b.String())
		}
	}
}
