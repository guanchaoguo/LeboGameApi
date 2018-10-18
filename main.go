package main

import (
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
	"golang-LeboGameApi/app/helper"
	"golang-LeboGameApi/routes"
	"net/http"
	_ "net/http/pprof"
	"strings"
	//"log"
	"github.com/iris-contrib/middleware/cors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
]
  
func newApp() *iris.Application {
	app := iris.New()

	//获取当前执行文件的路径
	file, _ := exec.LookPath(os.Args[0])
	AppPath, _ := filepath.Abs(file)
	losPath, _ := filepath.Split(AppPath)

	app.Get("/apidoc", func(ctx iris.Context) {
		ctx.ServeFile(losPath+"/apidoc/index.html", false)
	})

	fileServer := app.StaticHandler(losPath+"/apidoc", false, false)

	app.WrapRouter(func(w http.ResponseWriter, r *http.Request, router http.HandlerFunc) {
		path := r.URL.Path

		if !strings.Contains(path, ".") {
			router(w, r)
			return
		}
		ctx := app.ContextPool.Acquire(w, r)
		fileServer(ctx)
		app.ContextPool.Release(ctx)
	})

	return app
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU() * 2)
	app := newApp()
	app.WrapRouter(cors.WrapNext(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"*"},
		AllowedHeaders: []string{"*"},
	}))
	//go func() {
	//	log.Println(http.ListenAndServe("localhost:6060", nil))
	//}()
	//访问日志处理
	r, close := helper.NewRequestLogger()
	defer close()
	app.Use(r)
	app.OnAnyErrorCode(r, func(ctx iris.Context) {
		ctx.Writef("(Unexpected) internal server error")
	})

	// 错误处理
	app.OnErrorCode(iris.StatusInternalServerError, func(ctx context.Context) {
		errMessage := ctx.Values().GetString("error")
		if errMessage != "" {
			ctx.Writef("Internal server error: %s", errMessage)

			// 记录到日志文件
			helper.Self_logger(errMessage)
			return
		}

		ctx.Writef("(Unexpected) internal server error")
	})
	// 控制台输出程序启动信息
	//app.Use(func(this context.Context) {
	//	this.Application().Logger().Infof("Begin request for path %s", this.Path())
	//	this.Next()
	//})

	// 开启路由
	routes.Routes{}.WebRoute(app)

	// 执行监听
	app.Run(iris.Addr(":8099"), iris.WithoutVersionChecker, iris.WithCharset("UTF-8"),
		//获取真实ip配置
		iris.WithConfiguration(
			iris.Configuration{
				RemoteAddrHeaders: map[string]bool{
					"X-Real-Ip":             true,
					"X-Forwarded-For":       true,
					"CF-Connecting-IP": true,
				},
			},
			))
}
