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
    logType byte
)

const (
    log_none = logType(0x0)
    log_fatal = logType(0x1)
    log_error = logType(0x2)
    log_warn = logType(0x4)
    log_info = logType(0x8)
    log_debug = logType(0x10)
)

const (
    LOG_LEVEL_NONE = LogLevel(log_none)
    LOG_LEVEL_FATAL = LOG_LEVEL_NONE | LogLevel(log_fatal)
    LOG_LEVEL_ERROR = LOG_LEVEL_FATAL | LogLevel(log_error)
    LOG_LEVEL_WARN = LOG_LEVEL_ERROR | LogLevel(log_warn)
    LOG_LEVEL_INFO = LOG_LEVEL_WARN | LogLevel(log_info)
    LOG_LEVEL_DEBUG = LOG_LEVEL_INFO | LogLevel(log_debug)
    LOG_LEVEL_ALL = LOG_LEVEL_DEBUG
)

var _log *LevelLogger = New()

func LevLogger() *LevelLogger {
    return _log
}

func Logger() *log.Logger {
    return _log._log
}

func SetLogLevel(level string) {
    _log.SetLogLevel(level)
}
func GetLogLevel() LogLevel {
    return _log.GetLogLevel()
}

func SetWriter(out io.Writer) {
    _log.SetWriter(out)
}

func SetFlags(flags int) {
    _log.SetFlags(flags)
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

type LevelLogger struct {
    *levelLogger
    depth int
    tracelogs map[string]*levelLogger
}

type levelLogger struct {
    _log *log.Logger
    level LogLevel
}

func AddTraceLog(w io.Writer, level string) string {
    return _log.AddTraceLog(w, level)
}

func DelTraceLog(id string) {
    _log.DelTraceLog(id)
}

func (l *LevelLogger) AddTraceLog(w io.Writer, level string) string {
    buf := make([]byte, 20)
    io.ReadFull(rand.Reader, buf)
    name := hex.EncodeToString(buf)
    _log.tracelogs[name] = &levelLogger{log.New(w, l._log.Prefix(), l._log.Flags()), stringToLogLevel(level)}
    return name
}

func (l *LevelLogger) DelTraceLog(id string) {
    delete(l.tracelogs, id)
}

func (l *LevelLogger) SetLogLevel(level string) {
    l.level = stringToLogLevel(level)
}

func (l *LevelLogger) GetLogLevel() LogLevel {
    return l.level
}

func (l *LevelLogger) SetFlags(flags int) {
    l._log.SetFlags(flags)
}

func (l *LevelLogger) SetDepth(depth int) {
    l.depth = depth
}

func (l *LevelLogger) log(t logType, v ...interface{}) {
    if l.level | LogLevel(t) != l.level {
        return
    }

    s := l.convert2string(t, v...)
    l._log.Output(l.depth, s)

    var invalidTraceLogs []string
    for name, hk := range l.tracelogs {
        if hk.level | LogLevel(t) != hk.level {
            continue
        }

        err := hk._log.Output(l.depth, s)
        if err != nil {
            l._log.Output(l.depth, l.convert2string(log_error, "Error when write log to hook logs:", err))
            invalidTraceLogs = append(invalidTraceLogs, name)
            continue
        }
    }

    for _, name := range invalidTraceLogs {
        delete(l.tracelogs, name)
    }
}

func (*LevelLogger) convert2string(t logType, v ...interface{}) string {
    v1 := make([]interface{}, len(v)+2)
    logStr, logColor := logTypeToString(t)
    v1[0] = "\033" + logColor + "m[" + logStr + "]"
    copy(v1[1:], v)
    v1[len(v)+1] = "\033[0m"
    s := fmt.Sprintln(v1...)
    return s
}

func (l *LevelLogger) Fatal(v ...interface{}) {
    l.log(log_fatal, v...)
    os.Exit(-1)
}

func (l *LevelLogger) Error(v ...interface{}) {
    l.log(log_error, v...)
}

func (l *LevelLogger) Warn(v ...interface{}) {
    l.log(log_warn, v...)
}

func (l *LevelLogger) Debug(v ...interface{}) {
    l.log(log_debug, v...)
}

func (l *LevelLogger) Info(v ...interface{}) {
    l.log(log_info, v...)
}

func (l *LevelLogger) SetWriter(w io.Writer) {
    l._log = log.New(w, l._log.Prefix(), l._log.Flags())
}

func stringToLogLevel(level string) LogLevel {
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

func logTypeToString(t logType) (string, string) {
    switch t {
        case log_fatal:
            return "fatal", "[0;31"
        case log_error:
            return "error", "[0;31"
        case log_warn:
            return "warning", "[0;33"
        case log_debug:
            return "debug", "[0;36"
        case log_info:
            return "info", "[0;37"
    }
    return "unknown", "[0;37"
}

func New() *LevelLogger {
    return Newlogger(os.Stdout, "")
}

func Newlogger(w io.Writer, prefix string) *LevelLogger {
    return &LevelLogger{&levelLogger{log.New(w, prefix, LstdFlags), LOG_LEVEL_ALL}, 4, map[string]*levelLogger{}}
}
