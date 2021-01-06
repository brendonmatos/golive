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
	l.Prefix = aurora.BrightBlue("GL ").String()
	l.TimeFormat = aurora.BrightBlack("15:04:05").String()

	return &l
}

func (l *LoggerBasic) Log(level int, message string, extra map[string]interface{}) {
	// Level filter with override for fatal and panic
	if level < l.Level && level != LogFatal && level != LogPanic {
		return
	}

	b := strings.Builder{}
	b.WriteString(l.Prefix)

	b.WriteString(time.Now().Format(l.TimeFormat))

	switch level {
	case LogTrace:
		b.WriteString(aurora.Magenta(" TRC ").String())
	case LogDebug:
		b.WriteString(aurora.Yellow(" DBG ").String())
	case LogInfo:
		b.WriteString(aurora.Green(" INF ").String())
	case LogWarn:
		b.WriteString(aurora.Red(" WRN ").String())
	case LogError:
		b.WriteString(aurora.Red(" ERR ").Bold().String())
	case LogFatal:
		b.WriteString(aurora.Red(" FTL ").Bold().String())
	case LogPanic:
		b.WriteString(aurora.Red(" PNC ").Bold().String())
	default:
		b.WriteString(aurora.Bold(" ??? ").String())
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
		fmt.Println(b.String())
	}
}
