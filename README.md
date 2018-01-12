applog designed by golang

###日志记录服务, 可以定时切换日志文件

引用定时服务包 "github.com/robfig/cron"

###Example
import "applog"

logger = applog.NewAutoDailyLoger("D:\log", "log", "info")
logger.Start()
defer logger.Stop()