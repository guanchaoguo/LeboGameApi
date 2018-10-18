package routes

import (
	"github.com/kataras/iris/context"
	"github.com/kataras/iris/context"
	"golang-LeboGameApi/app/controllers"
)

type Routes struct{}

func (Routes) WebRoute(app *iris.Application) {

	app.Get("/", func(ctx context.Context) {
		ctx.Writef("Welcome to visit Lebo open platform!\n 歡迎訪問利博遊戲開放平台")
	})

	app.Get("/test", controllers.CashController{}.Test)

	/**
	 * 接入商获取SecurityKey
	 */
	app.Post("getKey", controllers.Key{}.Index)

	/**
	 * 会员充值
	 */
	app.Post("deposit", controllers.CashController{}.Deposit)
	/**
	 * 会员充值
	 */
	app.Post("deposit2", controllers.CashController{}.Test)
	/**
	 * 会员取款
	 */
	app.Post("withDrawal",controllers.CashController{}.WithDrawal)

	/**
	 * 会员存取款状态查询
	 */
	app.Post("transferLog", controllers.TransferLogController{}.Index)

	/**
	 * 会员登录&注册
	 */
	app.Post("authorization", controllers.AuthController{}.Login)

	/**
	 * 获取供应商会员信息
	 */
	app.Post("user", controllers.PlayerController{}.Index)

	/**
	 * 代理商批量获取注单信息
	 */
	app.Post("orderList", controllers.UserCash{}.OrderList)

	/**
	 * 代理商根据时间段获取注单信息
	 */
	app.Post("getDateList",controllers.UserCash{}.GetOrderListByDate)

	/**
	 * 异常注单信息获取
	 */
	app.Post("exceptionOrderLog", controllers.ExceptionOrderLogController{}.Index)

	/**
	 * 测试获取代理商信息
	 */
	app.Get("agent", controllers.PlayerController{}.Agent)

	/**
	 * 获取注单详情信息（代理商为联调账号时可用）
	 */
	app.Post("getBetList", controllers.UserCash{}.GetBetList)

}
