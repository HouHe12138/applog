package applog

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	//"github.com/Sirupsen/logrus"
	"github.com/robfig/cron"
)

const (
	SpecLog       = "@daily"
	SpecLogHours  = "@hourly"
	SpecLogMinute = "@every 1m"
	SpecLogSecond = "@every 10s"
)

//Auto Daily Save Log Manager
type AutoDailyLoger struct {
	dir    string
	prefix string
	file   *os.File
	cron   *cron.Cron
	level  string
}

func NewAutoDailyLoger(dir string, prefix string, level string) *AutoDailyLoger {
	c := cron.New()
	//init output 2006-01-02 15:04:05
	name := fmt.Sprintf("%v.log", filepath.Join(dir, prefix+time.Now().Format("2006010215")))
	fmt.Println("dir = ", dir, " ,name = ", name)
	os.MkdirAll(dir, 0777)
	file, _ := os.OpenFile(name, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0755)

	if file != nil {
		//log.SetOutput(file)
		//log.SetFlags(log.Llongfile | log.LstdFlags)
		SetOutput(file)
		lvl, err := ParseLevel(level)
		if err == nil {
			SetLevel(lvl)
		}
		SetFlags(Llongfile | LstdFlags)
	}

	//if file != nil {
	//	logrus.SetOutput(file)
	//	lvl, err := logrus.ParseLevel(level)
	//	if err == nil {
	//		logrus.SetLevel(lvl)
	//	}
	//	format := new(logrus.TextFormatter)
	//	format.DisableColors = true
	//	logrus.SetFormatter(format)
	//}
	//logrus.Info("NewAutoDailyLoger OK!!!")
	s := &AutoDailyLoger{
		dir:    dir,
		prefix: prefix,
		cron:   c,
		file:   file,
	}
	c.AddFunc(SpecLogHours, s.changeLogerFile)
	return s
}

func (s *AutoDailyLoger) Start() {
	s.cron.Start()
	Info("AutoDailyLoger start")
}

func (s *AutoDailyLoger) Stop() {
	s.cron.Stop()
	Info("AutoDailyLoger stop")
	if s.file != nil {
		s.file.Close()
		s.file = nil
	}
}

func (s *AutoDailyLoger) changeLogerFile() {
	//r std.mu.Unlock()
	if s.file != nil {
		s.file.Close()
		s.file = nil
	}
	name := fmt.Sprintf("%v.log", filepath.Join(s.dir, s.prefix+time.Now().Format("2006010215")))
	os.MkdirAll(s.dir, 0777)
	file, err := os.OpenFile(name, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0755)
	fmt.Println("file error", file, file.Name(), err)
	if file != nil {
		//log.SetOutput(s.file)
		//log.SetOutput(file)
		//log.SetFlags(log.Llongfile | log.LstdFlags)
		SetOutput(file)
		s.file = file
		lvl, err := ParseLevel(s.level)
		if err == nil {
			SetLevel(lvl)
		}
		SetFlags(Llongfile | LstdFlags)
		//lvl, err := logrus.ParseLevel(s.level)
		//if err == nil {
		//	logrus.SetLevel(lvl)
		//}
		//format := new(logrus.TextFormatter)
		//format.DisableColors = true
		//logrus.SetFormatter(format)
	}
	Info("changeLogerFile OK!!!")
}
