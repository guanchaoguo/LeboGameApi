package helper

import (
	"fmt"
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const deleteFileOnExit = false

//日志处理
// get a filename based on the date, file logs works that way the most times
// but these are just a sugar.
func todayFilename() string {
	today := time.Now().Format("2006-01-02")
	return today + ".log"
}

func newLogFile() *os.File {
	//获取当前执行文件的路径
	file, _ := exec.LookPath(os.Args[0])
	AppPath, _ := filepath.Abs(file)
	losPath, _ := filepath.Split(AppPath)
	//losPath,_ := os.Getwd()
	filename := todayFilename()
	path_d := losPath + "/logs/access/"
	if !isDirExists(path_d) {
		fmt.Println("目录不存在")
		if err := os.MkdirAll(path_d, 0777); err != nil {
			fmt.Printf("%s", err)
		} else {
			fmt.Print("Create Directory OK!")
		}
	}
	// open an output file, this will append to the today's file if server restarted.
	f, err := os.OpenFile(path_d+filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}

	return f
}

/**
 * 判断目录是否存在
 */
func isDirExists(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	} else {
		return fi.IsDir()
	}

	panic("not reached")
}

/**
 * 写日志文件
 */
func Self_logger(myerr interface{}) {
	logfile, err := os.OpenFile(newErrorLogFile(), os.O_CREATE|os.O_APPEND|os.O_RDWR, 0)
	if err != nil {
		fmt.Printf("%s\r\n", err.Error())
		os.Exit(-1)
	}
	defer logfile.Close()
	logger := log.New(logfile, "\r\n", log.Ldate|log.Ltime|log.Llongfile)
	logger.Println(myerr)

}
func newErrorLogFile() string {
	losPath, _ := os.Getwd()
	filename := todayFilename()
	path := losPath + "/logs/error/" + filename
	return path
}

var excludeExtensions = [...]string{
	".js",
	".css",
	".jpg",
	".png",
	".ico",
	".svg",
}

func NewRequestLogger() (h iris.Handler, close func() error) {
	close = func() error { return nil }

	c := logger.Config{
		Status:  true,
		IP:      true,
		Method:  true,
		Path:    true,
		Columns: true,
	}

	logFile := newLogFile()
	close = func() error {
		err := logFile.Close()
		if deleteFileOnExit {
			err = os.Remove(logFile.Name())
		}
		return err
	}

	c.LogFunc = func(now time.Time, latency time.Duration, status, ip, method, path string, message interface{}) {
		output := logger.Columnize(now.Format("2006/01/02 - 15:04:05"), latency, status, ip, method, path, message)
		logFile.Write([]byte(output))
	}

	//	we don't want to use the logger
	// to log requests to assets and etc
	c.AddSkipper(func(ctx iris.Context) bool {
		path := ctx.Path()
		for _, ext := range excludeExtensions {
			if strings.HasSuffix(path, ext) {
				return true
			}
		}
		return false
	})

	h = logger.New(c)

	return
}
