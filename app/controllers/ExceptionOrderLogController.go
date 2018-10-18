package controllers

import (
	"encoding/json"
	"github.com/garyburd/redigo/redis"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
	"golang-LeboGameApi/app/helper"
	"golang-LeboGameApi/app/middleware"
	"golang-LeboGameApi/app/models"
	"golang-LeboGameApi/config"
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"time"
)

/*
  异常注单控制器
*/

type ExceptionOrderLogController struct {
}

/**
 * @api {post} /exceptionOrderLog 异常注单信息获取
 * @apiDescription 异常注单信息获取
 * @apiGroup User
 * @apiPermission
 * @apiVersion 1.0.0
 * @apiParam {String} agent 代理商用户名
 * @apiParam {String} username 用户登录名称
 * @apiParam {String} start_date 开始时间
 * @apiParam {String} end_date  结束时间
 * @apiParam {String} token token:SHA1('securityKey|start_date|end_date|agent')
 * @apiSuccessExample {json} 成功返回格式
	{
		"code": 0,
		"result": {
			"data": [
				{
					"_id": "59fad2fce138230c391efe65",
					"user_name": "D01test12345",
					"hall_name": "tlebo01",
					"round_no": "2e0fac7f9449eeb0",
					"payout_win": -40,
					"user_monry": 0,
					"action_user": "chensj",
					"action_user_id": 0,
					"desc": "数组数组数组数组数组数组数组数组数组数组数(取消)",
					"bet_time": "2017-11-02 15:17:52",
					"add_time": "2017-11-02 16:10:36"
				}
			]
		},
		"text": "Success"
	}
*/
func (this ExceptionOrderLogController) Index(ctx context.Context) {
	//白名单验证
	middleware.IpLimit(ctx)

	//type Apilog struct {
	//	Agent      string
	//	Start_time time.Time
	//	End_time   time.Time
	//	PostData   []string
	//	ApiName    string
	//	Ip_info    string
	//}

	agent := ctx.FormValue("agent")
	start_date := ctx.FormValue("start_date")
	end_date := ctx.FormValue("end_date")
	token := ctx.FormValue("token")
	is_test := ctx.FormValue("is_test") //压测使用

	postData, _ := json.Marshal(ctx.FormValues())

	//记录调用日志
	apiLog := map[string]interface{}{
		"start_time": time.Now(),
		"agent":      agent,
		"postData":   string(postData),
		"apiName":    "时间段获取异常注单信息",
		"ip_info":    ctx.RemoteAddr(),
		"log_type":   "api",
	}
	//联调状态统计 只要有联调请求则统计数据
	is_debug := models.StatisticsLog{}.ApiStatistics(apiLog["apiName"].(string), agent)

	if is_test == "" {
		//判断调用频率间隔时间
		has, errCode := this.getListFeed(agent)
		if !has {
			apiLog["code"] = errCode
			apiLog["text"] = config.ErrCode(errCode)
			apiLog["result"] = ""
			apiLog["end_time"] = time.Now()
			go models.ApiLog{}.Insert(&apiLog)
			ctx.JSON(iris.Map{
				"code":   errCode,
				"text":   config.ErrCode(errCode),
				"result": "",
			})
			return
		}
	}

	//判断代理商名字和token是否为空
	if agent == "" || token == "" || start_date == "" || end_date == "" {
		ctx.JSON(iris.Map{
			"code":   9002,
			"text":   config.ErrCode(9002),
			"result": "",
		})
		return
	}

	//判断系统是否在维护当中
	has, err_Code := middleware.CheckIsMaintain(agent, "")
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
	//时间跨度认证 不能超过一周
	startDate, _ := time.Parse("2006-01-02 15:04:05", start_date)
	endDate, _ := time.Parse("2006-01-02 15:04:05", end_date)
	if endDate.Unix()-startDate.Unix() > 7*24*60*60 {
		ctx.JSON(iris.Map{
			"code":   2001,
			"text":   config.ErrCode(2001),
			"result": "",
		})
		return
	}

	//token认证对比
	key_res, has := GetSecurityKey(agent)
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
	checkEequest := helper.CheckEequest(token, []string{securityKey, start_date, end_date, agent})

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

	//根据代理商获取代理商ID
	user_info, err := models.Agent{}.GetAgentInfo(agent)

	if err != nil || user_info == nil {
		apiLog["code"] = 9004
		apiLog["text"] = config.ErrCode(9004)
		apiLog["result"] = ""
		apiLog["end_time"] = time.Now()
		go models.ApiLog{}.Insert(&apiLog)

		ctx.JSON(iris.Map{
			"code":   9004,
			"text":   config.ErrCode(9004),
			"result": "",
		})
		return
	}

	//验证通过进行数据获取操作
	list_data, err := this.getOrderList(user_info.Id, start_date, end_date)
	if err != nil || list_data == nil {
		apiLog["code"] = 9009
		apiLog["text"] = config.ErrCode(9009)
		apiLog["result"] = ""
		apiLog["end_time"] = time.Now()
		go models.ApiLog{}.Insert(&apiLog)
		ctx.JSON(iris.Map{
			"code":   9009,
			"text":   config.ErrCode(9009),
			"result": "",
		})
		return
	}

	//联调账号联调成功次数统计
	if is_debug {
		go models.StatisticsLog{}.ApiSucceds(apiLog["apiName"].(string), agent)
	}

	apiLog["code"] = 0
	apiLog["text"] = config.ErrCode(0)
	apiLog["result"] = ""
	apiLog["end_time"] = time.Now()
	go models.ApiLog{}.Insert(&apiLog)

	ctx.JSON(iris.Map{
		"code": 0,
		"text": config.ErrCode(9999),
		"result": iris.Map{
			"data": list_data,
		},
	})
	//ctx.JSON(iris.Map{
	//	"code":   100,
	//	"text":   config.ErrCode(9002),
	//	"result": "ok",
	//})
	//return
}

/**
 * 监测时间间隔
 */
func (ExceptionOrderLogController) getListFeed(agent_name string) (bool, int) {
	keyName := agent_name + config.Common().AGENT_ORDER_LIST
	c := helper.RedisClientDefault.Get()
	defer c.Close()
	lastDate, _ := redis.String(c.Do("GET", keyName))
	lastTime, _ := strconv.ParseInt(lastDate, 10, 64)
	limit_time := config.Common().GET_ORDER_LIST_SPEED

	//判断是否在有效时间间隔内
	if lastTime > 0 && time.Now().Unix()-lastTime < limit_time {
		return false, 9008
	}

	//符合速率则刷新本次调用时间
	redis.String(c.Do("SET", keyName, time.Now().Unix()))

	return true, 0
}

/**
 * 获取注单列表
 */
func (ExceptionOrderLogController) getOrderList(agent_id int, startDate string, endDate string) ([]models.ExceptionOrderLog, error) {

	//获取单次最大获取的数据量条数，可配置
	take := config.Common().GET_ORDER_LIST_MAX

	//根据最大ID获取数据
	findWhere := bson.M{}
	//if len(deadline) > 0 {
	//	deadline_utc, _ := time.ParseInLocation("2006-01-02 15:04:05 ", deadline, time.Local)
	//	//findWhere = bson.M{ "agent_id": agent_id,  "is_mark":1 ,"end_time":bson.M{"$gte":deadline_utc} }
	//	findWhere = bson.M{"agent_id": agent_id, "is_mark": 1, "start_time": bson.M{"$gte": deadline_utc}}
	//}

	//根据时间段获取数据
	if len(startDate) > 0 && len(endDate) > 0 {
		startDate_utc, _ := time.ParseInLocation("2006-01-02 15:04:05 ", startDate, time.Local)
		endDate_utc, _ := time.ParseInLocation("2006-01-02 15:04:05 ", endDate, time.Local)
		//findWhere = bson.M{"agent_id": agent_id, "is_mark":1 , "end_time":bson.M{"$gte":startDate_utc, "$lte":endDate_utc} }
		findWhere = bson.M{"agent_id": agent_id, "add_time": bson.M{"$gte": startDate_utc, "$lte": endDate_utc}}
	}
	//fmt.Println(findWhere)
	// 获取的数据 字段
	pipe := []bson.M{
		{"$match": findWhere},
		{"$project": bson.M{
			"action_user": 1,
			"bet_time":    1,
			"add_time":    1,
			"desc":        1,
			"hall_name":   1,
			"user_name":   1,
			"payout_win":  1,
			"round_no":    1,
			"user_money":  1,
		},
		},
		{"$limit": take},
	}

	user_chart, err := models.ExceptionOrderLog{}.GetListData(pipe, take)
	//fmt.Println(user_chart_Info)
	if err != nil || len(user_chart) == 0 {
		return nil, err
	}
	return user_chart, nil
}
