//high level log wrapper, so it can output different log based on level
package levelog

import (
    "io"
    "os"
    "log"
    "fmt"
    "crypto/rand"
    "encoding/hex"
)

const (
    Ldate = log.Ldate
    Llongfile = log.Llongfile
    Lmicroseconds = log.Lmicroseconds
    Lshortfile = log.Lshortfile
    LstdFlags = log.LstdFlags
    Ltime = log.Ltime
)

type (
    LogLevel byte
    LogType byte
)

const (
    LOG_FATAL = LogType(0x1)
    LOG_ERROR = LogType(0x2)
    LOG_WARN = LogType(0x4)
    LOG_DEBUG = LogType(0x8)
    LOG_INFO = LogType(0x10)
)

const (
    LOG_LEVEL_NONE = LogLevel(0x0)
    LOG_LEVEL_FATAL = LOG_LEVEL_NONE | LogLevel(LOG_FATAL)
    LOG_LEVEL_ERROR = LOG_LEVEL_FATAL | LogLevel(LOG_ERROR)
    LOG_LEVEL_WARN = LOG_LEVEL_ERROR | LogLevel(LOG_WARN)
    LOG_LEVEL_DEBUG = LOG_LEVEL_WARN | LogLevel(LOG_DEBUG)
    LOG_LEVEL_INFO = LOG_LEVEL_DEBUG | LogLevel(LOG_INFO)
    LOG_LEVEL_ALL = LOG_LEVEL_INFO
)

var _log *logger = New()

func Logger() *log.Logger {
    return _log._log
}

func SetLogLevel(level string) {
    _log.SetLogLevel(level)
}
func GetLogLevel() LogLevel {
    return _log.level
}

func SetWriter(out io.Writer) {
    _log.SetWriter(out)
}

func SetFlags(flags int) {
    _log._log.SetFlags(flags)
}

func Info(v ...interface{}) {
    _log.Info(v...)
}

func Debug(v ...interface{}) {
    _log.Debug(v...)
}

func Warn(v ...interface{}) {
    _log.Warn(v...)
}

func Error(v ...interface{}) {
    _log.Error(v...)
}

func Fatal(v ...interface{}) {
    _log.Fatal(v...)
}

type logger struct {
    _log *log.Logger
    level LogLevel
    tracelogs map[string]*HookLogger
}

type HookLogger struct{
    level LogLevel
    _log *log.Logger
    WriteFlusher
}

func NewHookLog(wf WriteFlusher, level LogLevel) *HookLogger {
    return &HookLogger{level, log.New(wf, _log._log.Prefix(), _log._log.Flags()), wf}
}

type WriteFlusher interface {
    Write([]byte) (int, error)
    Flush()
}

func NewTraceLog(wf WriteFlusher) string {
    return NewTraceLogHookLog(NewHookLog(wf, _log.level))
}

func NewTraceLogHookLog(hl *HookLogger) string {
    buf := make([]byte, 20)
    io.ReadFull(rand.Reader, buf)
    name := hex.EncodeToString(buf)
    _log.tracelogs[name] = hl
    return name
}

func (l *logger) SetLogLevel(level string) {
    l.level = StringToLogLevel(level)
}

func (l *logger) log(t LogType, v ...interface{}) {
    if l.level | LogLevel(t) != l.level {
        return
    }

    s := l.convert2string(t, v...)
    l._log.Output(4, s)

    var invalidTraceLogs []string
    for name, hk := range l.tracelogs {
        if hk.level | LogLevel(t) != hk.level {
            continue
        }

        err := hk._log.Output(4, s)
        if err != nil {
            l._log.Output(4, l.convert2string(LOG_ERROR, "Error when write log to hook logs:", err))
            invalidTraceLogs = append(invalidTraceLogs, name)
            continue
        }
        hk.Flush()
    }

    for _, name := range invalidTraceLogs {
        delete(l.tracelogs, name)
    }
}

func (*logger) convert2string(t LogType, v ...interface{}) string {
    v1 := make([]interface{}, len(v)+2)
    logStr, logColor := LogTypeToString(t)
    v1[0] = "\033" + logColor + "m[" + logStr + "]"
    copy(v1[1:], v)
    v1[len(v)+1] = "\033[0m"
    s := fmt.Sprintln(v1...)
    return s
}

func (l *logger) Fatal(v ...interface{}) {
    l.log(LOG_FATAL, v...)
    os.Exit(-1)
}

func (l *logger) Error(v ...interface{}) {
    l.log(LOG_ERROR, v...)
}

func (l *logger) Warn(v ...interface{}) {
    l.log(LOG_WARN, v...)
}

func (l *logger) Debug(v ...interface{}) {
    l.log(LOG_DEBUG, v...)
}

func (l *logger) Info(v ...interface{}) {
    l.log(LOG_INFO, v...)
}

func (l *logger) SetWriter(w io.Writer) {
    l._log = log.New(w, l._log.Prefix(), l._log.Flags())
}

func StringToLogLevel(level string) LogLevel {
    switch level {
        case "fatal":
            return LOG_LEVEL_FATAL
        case "error":
            return LOG_LEVEL_ERROR
        case "warn":
            return LOG_LEVEL_WARN
        case "debug":
            return LOG_LEVEL_DEBUG
        case "info":
            return LOG_LEVEL_INFO
    }
    return LOG_LEVEL_ALL
}

func LogTypeToString(t LogType) (string, string) {
    switch t {
        case LOG_FATAL:
            return "fatal", "[0;31"
        case LOG_ERROR:
            return "error", "[0;31"
        case LOG_WARN:
            return "warning", "[0;33"
        case LOG_DEBUG:
            return "debug", "[0;36"
        case LOG_INFO:
            return "info", "[0;37"
    }
    return "unknown", "[0;37"
}

func New() *logger {
    return Newlogger(os.Stdout, "")
}

func Newlogger(w io.Writer, prefix string) *logger {
    return &logger{log.New(w, prefix, LstdFlags), LOG_LEVEL_ALL, map[string]*HookLogger{}}
}
