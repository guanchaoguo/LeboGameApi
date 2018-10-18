package controllers

import (
	"encoding/json"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
	"golang-LeboGameApi/app/helper"
	"golang-LeboGameApi/app/middleware"
	"golang-LeboGameApi/app/models"
	"golang-LeboGameApi/config"
	"time"
)

/**
 * 会员存取款状态查询
 */
type TransferLogController struct {
}

/**
 * @api {post} /transferLog 会员存取款状态查询
 * @apiDescription  当游戏接入商调用会员存款或取款接口，数据已发出但未接收到处理结果，
					则可使用此接口至游戏供应商查询处理状态。目前只支持单条记录查询。
 * @apiGroup User
 * @apiPermission
 * @apiVersion 1.0.0
 * @apiParam {String} agent 代理商用户名
 * @apiParam {String} serial 流水号
 * @apiParam {String} token token:SHA1('securityKey|serial|agent')
 * @apiSuccessExample {json} 成功返回格式
	{
		"code": 0,//状态码，0：成功，非0：错误
		"result": {
			"balance": 1332234,//余额
			"member_status": "Normal",//会员状态，Normal：正常，Abnormal：异常
			"online_status": "Offline"//在线状态，Online：在线，Offline：离线
		},
		"text": "Success"//文本描述
	}
*/
func (TransferLogController) Index(ctx context.Context) {
	//白名单验证
	if res := middleware.IpLimit(ctx); !res{
		return
	}

	agent_name := ctx.FormValue("agent")
	serial := ctx.FormValue("serial")
	token := ctx.FormValue("token")

	postData, _ := json.Marshal(ctx.FormValues())
	//记录调用日志
	apiLog := map[string]interface{}{
		"start_time": time.Now(),
		"user_name":  agent_name,
		"postData":   string(postData),
		"apiName":    "会员存取款状态查询",
		"ip_info":    ctx.RemoteAddr(),
		"log_type":   "api",
	}

	//联调状态统计 只要有联调请求则统计数据
	is_debug := models.StatisticsLog{}.ApiStatistics(apiLog["apiName"].(string), agent_name)

	//判断系统是否在维护当中
	has, err_Code := middleware.CheckIsMaintain(agent_name, "")
	if !has {
		apiLog["code"] = err_Code
		apiLog["text"] = config.ErrCode(err_Code)
		apiLog["result"] = ""
		apiLog["end_time"] = time.Now()
		go models.ApiLog{}.Insert(&apiLog)
		ctx.JSON(iris.Map{
			"code":   err_Code,
			"text":   config.ErrCode(err_Code),
			"result": "",
		})
		return
	}

	if agent_name == "" || serial == "" || token == "" {
		apiLog["code"] = 9002
		apiLog["text"] = config.ErrCode(9002)
		apiLog["result"] = ""
		apiLog["end_time"] = time.Now()
		go models.ApiLog{}.Insert(&apiLog)

		ctx.JSON(iris.Map{
			"code":   9002,
			"text":   config.ErrCode(9002),
			"result": "",
		})
		return
	}

	//token认证对比
	key_res, has := GetSecurityKey(agent_name)
	if !has {
		err_code := key_res.(int)
		apiLog["code"] = key_res
		apiLog["text"] = config.ErrCode(err_code)
		apiLog["result"] = ""
		apiLog["end_time"] = time.Now()
		go models.ApiLog{}.Insert(&apiLog)
		ctx.JSON(iris.Map{
			"code":   key_res,
			"text":   config.ErrCode(err_code),
			"result": "",
		})
		return
	}
	securityKey := key_res.(string)
	checkEequest := helper.CheckEequest(token, []string{securityKey, serial, agent_name})

	if !checkEequest {

		apiLog["code"] = 9002
		apiLog["text"] = config.ErrCode(9002)
		apiLog["result"] = ""
		apiLog["end_time"] = time.Now()
		go models.ApiLog{}.Insert(&apiLog)

		ctx.JSON(iris.Map{
			"code":   9002,
			"text":   config.ErrCode(9002),
			"result": "",
		})
		return
	}

	pkey := helper.MD5(agent_name + serial + config.Common().GAME_API_SUF)
	res, err := models.CashRecord{}.GetInfoByPkey(pkey)
	//数据不存在
	if res.Order_sn != serial || err != nil {
		apiLog["code"] = 1008
		apiLog["text"] = config.ErrCode(1008)
		apiLog["result"] = ""
		apiLog["end_time"] = time.Now()
		go models.ApiLog{}.Insert(&apiLog)

		ctx.JSON(iris.Map{
			"code":   1008,
			"text":   config.ErrCode(1008),
			"result": "",
		})
		return
	}

	apiLog["code"] = 0
	apiLog["text"] = config.ErrCode(9999)
	apiLog["result"] = ""
	apiLog["end_time"] = time.Now()
	go models.ApiLog{}.Insert(&apiLog)

	//联调账号联调成功次数统计
	if is_debug {
		go models.StatisticsLog{}.ApiSucceds(apiLog["apiName"].(string), agent_name)
	}

	ctx.JSON(iris.Map{
		"code":   0,
		"text":   config.ErrCode(9999),
		"result": "",
	})
}
