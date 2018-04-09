package applog

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"
)

const CallPath = 3

type Logger struct {
	// The logs are `io.Copy`'d to this in a mutex. It's common to set this to a
	// file, or leave it default which is `os.Stderr`. You can also set this to
	// something more adventorous, such as logging to Kafka.
	Out io.Writer
	// Used to sync writing to the log. Locking is enabled by Default
	mu sync.Mutex
	// The logging level the logger should log at. This is typically (and defaults
	// to) `logrus.Info`, which allows Info(), Warn(), Error() and Fatal() to be
	// logged. `logrus.Debug` is useful in
	Level Level
	// properties
	flag int
	// for accumulating text to write
	buf []byte
}

type MutexWrap struct {
	lock     sync.Mutex
	disabled bool
}

func (mw *MutexWrap) Lock() {
	if !mw.disabled {
		mw.lock.Lock()
	}
}

func (mw *MutexWrap) Unlock() {
	if !mw.disabled {
		mw.lock.Unlock()
	}
}

func (mw *MutexWrap) Disable() {
	mw.disabled = true
}

// It's recommended to make this a global instance called `log`.
func New() *Logger {
	return &Logger{
		Out:   os.Stderr,
		Level: InfoLevel,
		flag:  LstdFlags,
	}
}

// Cheap integer to fixed-width decimal ASCII.  Give a negative width to avoid zero-padding.
func itoa(buf *[]byte, i int, wid int) {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}

func (l *Logger) formatHeader(buf *[]byte, t time.Time, file string, line int) {
	//*buf = append(*buf, l.prefix...)
	if l.flag&LUTC != 0 {
		t = t.UTC()
	}
	if l.flag&(Ldate|Ltime|Lmicroseconds) != 0 {
		if l.flag&Ldate != 0 {
			year, month, day := t.Date()
			itoa(buf, year, 4)
			*buf = append(*buf, '/')
			itoa(buf, int(month), 2)
			*buf = append(*buf, '/')
			itoa(buf, day, 2)
			*buf = append(*buf, ' ')
		}
		if l.flag&(Ltime|Lmicroseconds) != 0 {
			hour, min, sec := t.Clock()
			itoa(buf, hour, 2)
			*buf = append(*buf, ':')
			itoa(buf, min, 2)
			*buf = append(*buf, ':')
			itoa(buf, sec, 2)
			if l.flag&Lmicroseconds != 0 {
				*buf = append(*buf, '.')
				itoa(buf, t.Nanosecond()/1e3, 6)
			}
			*buf = append(*buf, ' ')
		}
	}
	if l.flag&(Lshortfile|Llongfile) != 0 {
		if l.flag&Lshortfile != 0 {
			short := file
			for i := len(file) - 1; i > 0; i-- {
				if file[i] == '/' {
					short = file[i+1:]
					break
				}
			}
			file = short
		}
		*buf = append(*buf, file...)
		*buf = append(*buf, ':')
		itoa(buf, line, -1)
		*buf = append(*buf, ": "...)
	}
}

func (l *Logger) Output(calldepth int, s string) error {
	now := time.Now()
	var file string
	var line int
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.flag&(Lshortfile|Llongfile) != 0 {
		// release lock while getting caller info - it's expensive.
		l.mu.Unlock()
		var ok bool
		_, file, line, ok = runtime.Caller(calldepth)
		if !ok {
			file = "???"
			line = 0
		}
		l.mu.Lock()
	}
	l.buf = l.buf[:0]
	l.formatHeader(&l.buf, now, file, line)
	l.buf = append(l.buf, s...)
	if len(s) == 0 || s[len(s)-1] != '\n' {
		l.buf = append(l.buf, '\n')
	}
	_, err := l.Out.Write(l.buf)
	return err
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	if l.Level >= DebugLevel {
		//fmt.Fprintf(l.Out, "[level=debug] "+format+"%s", args, "\n")
		l.Output(CallPath, fmt.Sprintf("[level=debug] "+format, args...)+"\n")
		//log.Printf("[level=debug] "+format, args, "\n")
	}
}

func (l *Logger) Infof(format string, args ...interface{}) {
	if l.Level >= InfoLevel {
		//fmt.Fprintf(l.Out, "[level=info] "+format+"%s", args, "\n")
		l.Output(CallPath, fmt.Sprintf("[level=info] "+format, args...)+"\n")
	}
}

func (l *Logger) Printf(format string, args ...interface{}) {
	//fmt.Fprintf(l.Out, format+"%s", args, "\n")
	l.Output(CallPath, fmt.Sprintf(format+"%s", args...)+"\n")
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	if l.Level >= WarnLevel {
		//fmt.Fprintf(l.Out, "[level=warning] "+format+"%s", args, "\n")
		l.Output(CallPath, fmt.Sprintf("[level=warning] "+format, args...)+"\n")
	}
}

func (l *Logger) Warningf(format string, args ...interface{}) {
	if l.Level >= WarnLevel {
		//fmt.Fprintf(l.Out, "[level=warning] "+format+"%s", args, "\n")
		l.Output(CallPath, fmt.Sprintf("[level=warning] "+format, args...)+"\n")
	}
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	if l.Level >= ErrorLevel {
		//fmt.Fprintf(l.Out, "[level=error] "+format+"%s", args, "\n")
		l.Output(CallPath, fmt.Sprintf("[level=error] "+format, args...)+"\n")
	}
}

func (l *Logger) Debug(args ...interface{}) {
	if l.Level >= DebugLevel {
		//fmt.Fprint(l.Out, "[level=debug] ", args, "\n")
		l.Output(CallPath, "[level=debug] "+fmt.Sprintln(args...))
	}
}

func (l *Logger) Info(args ...interface{}) {
	if l.Level >= InfoLevel {
		//fmt.Fprint(l.Out, "[level=info] ", args, "\n")
		l.Output(CallPath, "[level=info] "+fmt.Sprintln(args...))
	}
}

func (l *Logger) Print(args ...interface{}) {
	//fmt.Fprint(l.Out, args, "\n")
	l.Output(CallPath, fmt.Sprintln(args...))
}

func (l *Logger) Warn(args ...interface{}) {
	if l.Level >= WarnLevel {
		//fmt.Fprint(l.Out, "[level=warning] ", args, "\n")
		l.Output(CallPath, "[level=warning] "+fmt.Sprintln(args...))
	}
}

func (l *Logger) Warning(args ...interface{}) {
	if l.Level >= WarnLevel {
		//fmt.Fprint(l.Out, "[level=warning] ", args, "\n")
		l.Output(CallPath, "[level=warning] "+fmt.Sprintln(args...))
	}
}

func (l *Logger) Error(args ...interface{}) {
	if l.Level >= ErrorLevel {
		//fmt.Fprint(l.Out, "[level=error] ", args, "\n")
		l.Output(CallPath, "[level=error] "+fmt.Sprintln(args...))
	}
}

func (l *Logger) Debugln(args ...interface{}) {
	if l.Level >= DebugLevel {
		//fmt.Fprintln(l.Out, "[level=debug] ", args)
		l.Output(CallPath, "[level=debug] "+fmt.Sprintln(args...))
	}
}

func (l *Logger) Infoln(args ...interface{}) {
	if l.Level >= InfoLevel {
		//fmt.Fprintln(l.Out, "[level=info] ", args)
		l.Output(CallPath, "[level=info] "+fmt.Sprintln(args...))
	}
}

func (l *Logger) Println(args ...interface{}) {
	//fmt.Fprintln(l.Out, args)
	l.Output(CallPath, fmt.Sprintln(args...))
}

func (l *Logger) Warnln(args ...interface{}) {
	if l.Level >= WarnLevel {
		//fmt.Fprintln(l.Out, "[level=warning] ", args)
		l.Output(CallPath, "[level=warning] "+fmt.Sprintln(args...))
	}
}

func (l *Logger) Warningln(args ...interface{}) {
	if l.Level >= WarnLevel {
		//fmt.Fprintln(l.Out, "[level=warning] ", args)
		l.Output(CallPath, "[level=warning] "+fmt.Sprintln(args...))
	}
}

func (l *Logger) Errorln(args ...interface{}) {
	if l.Level >= ErrorLevel {
		//fmt.Fprintln(l.Out, "[level=error] ", args)
		l.Output(CallPath, "[level=error] "+fmt.Sprintln(args...))
	}
}
