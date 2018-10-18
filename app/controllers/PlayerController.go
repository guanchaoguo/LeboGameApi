package controllers

import (
	"encoding/json"
	"github.com/garyburd/redigo/redis"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
	"github.com/tidwall/gjson"
	"golang-LeboGameApi/app/helper"
	"golang-LeboGameApi/app/middleware"
	"golang-LeboGameApi/app/models"
	"golang-LeboGameApi/config"
	"golang-LeboGameApi/lib"
	"strconv"
	"time"
)

/**
玩家控制器
*/
type PlayerController struct{}

/**
 * @api {post} /user 获取供应商会员信息
 * @apiDescription 获取会员在供应商的相关信息
 * @apiGroup User
 * @apiPermission
 * @apiVersion 1.0.0
 * @apiParam {String} agent 代理商用户名
 * @apiParam {String} username 用户登录名称
 * @apiParam {String} token token:SHA1('securityKey|username|agent')
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
func (PlayerController) Index(ctx context.Context) {
	//白名单验证
	if res := middleware.IpLimit(ctx); !res{
		return
	}

	agent_name := ctx.FormValue("agent")
	user_name := ctx.FormValue("username")
	token := ctx.FormValue("token")

	postData, _ := json.Marshal(ctx.FormValues())
	//记录调用日志
	apiLog := map[string]interface{}{
		"start_time": time.Now(),
		"user_name":  agent_name,
		"postData":   string(postData),
		"apiName":    "获取供应商会员信息",
		"ip_info":    ctx.RemoteAddr(),
		"log_type":   "api",
	}

	//联调状态统计 只要有联调请求则统计数据
	is_debug := models.StatisticsLog{}.ApiStatistics(apiLog["apiName"].(string), agent_name)

	//判断系统是否在维护当中
	has, err_Code := middleware.CheckIsMaintain(agent_name, user_name)
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
	if agent_name == "" || user_name == "" || token == "" {

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
	checkEequest := helper.CheckEequest(token, []string{securityKey, user_name, agent_name})

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

	//获取代理商code
	redis_key := config.Common().WHITELIST
	c := helper.RedisClientDefault.Get()
	defer c.Close()
	redis_data, err := redis.String(c.Do("hget", redis_key, agent_name))
	agent_code := gjson.Get(redis_data, "agent_code").String()
	connect_mode := gjson.Get(redis_data, "connect_mode").Int()

	//会员名称加密
	decry_user_name := lib.Crypto{}.Decrypt(agent_code + user_name)

	res, err := models.User{}.GetInfo(decry_user_name)
	if err != nil {

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

	if res == nil {

		apiLog["code"] = 1007
		apiLog["text"] = config.ErrCode(1007)
		apiLog["result"] = ""
		apiLog["end_time"] = time.Now()
		go models.ApiLog{}.Insert(&apiLog)

		ctx.JSON(iris.Map{
			"code":   1007,
			"text":   config.ErrCode(1007),
			"result": "",
		})
		return
	}

	if  res.Account_state != 1{
		apiLog["code"] = 1001
		apiLog["text"] = config.ErrCode(1001)
		apiLog["result"] = ""
		apiLog["end_time"] = time.Now()
		go models.ApiLog{}.LoginLog(&apiLog)

		ctx.JSON(iris.Map{
			"code":   1001,
			"text":   config.ErrCode(1001),
			"result": "",
		})
		return
	}

	var online_status, member_status string

	if res.On_line == "Y" {
		online_status = "Online"
	} else {
		online_status = "Offline"
	}

	if res.Account_state == 1 {
		member_status = "Normal"
	} else {
		member_status = "Abnormal"
	}

	//调用日志记录
	result, _ := json.Marshal(iris.Map{
		"online_status": online_status,
		"member_status": member_status,
		"balance":       res.Money,
	})
	apiLog["code"] = 0
	apiLog["text"] = config.ErrCode(9999)
	apiLog["result"] = string(result)
	apiLog["end_time"] = time.Now()
	go models.ApiLog{}.Insert(&apiLog)

	//联调账号联调成功次数统计
	if is_debug {
		go models.StatisticsLog{}.ApiSucceds(apiLog["apiName"].(string), agent_name)
	}
	balance := "0.00"
	if connect_mode == 0 {
		balance = strconv.FormatFloat(res.Money, 'f', 2, 64)
	}
	ctx.JSON(iris.Map{
		"code": 0,
		"text": config.ErrCode(9999),
		"result": iris.Map{
			"online_status": online_status,
			"member_status": member_status,
			"balance":       balance,
		}})
}

func (PlayerController) Agent(ctx context.Context) {
	res, _ := models.Agent{}.GetAgentTest(2)
	ctx.JSON(iris.Map{
		"code":   0,
		"text":   config.ErrCode(9999),
		"result": res,
	})
}
