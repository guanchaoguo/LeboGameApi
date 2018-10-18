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

/**
 * 代理商获取注单信息
 */
type UserCash struct{}

/**
 * @api {post} /orderList 代理商批量获取注单信息
 * @apiDescription 代理商批量获取注单信息
 * @apiGroup User
 * @apiPermission
 * @apiVersion 1.0.0
 * @apiParam {String} agent 代理商用户名
 * @apiParam {String} username 用户登录名称
 * @apiParam {String} deadline 日期
 * @apiParam {String} token token:SHA1('securityKey|deadline|agent')
 * @apiSuccessExample {json} 成功返回格式
	{
		"code": 0,
		"result": {
			"data": [
				{
					"round_no": "3c4c0e4df67432bd",
					"game_id": 93,
					"game_hall_id": 0,
					"game_name": "龙虎 ",
					"server_name": "15",
					"total_bet_score": 70,
					"valid_bet_score_total": 70,
					"game_result": "8;5",
					"total_win_score": 70,
					"user_name": "test123451",
					"is_mark": 1,
					"dwRound": "",
					"remark": "24;53",
					"_id": "5a056e10e1382363171f7ec3",
					"game_period": "248-56",//靴+局
					"start_time": "2017-11-10 17:15:06"
				}
			]
		},
		"text": "Success"
	}
*/
func (this UserCash) OrderList(ctx context.Context) {
	//白名单验证
	if res := middleware.IpLimit(ctx); !res{
		return
	}

	agent_name := ctx.FormValue("agent")
	deadline := ctx.FormValue("deadline")
	token := ctx.FormValue("token")
	is_test := ctx.FormValue("is_test") //压测使用

	//记录调用日志
	postData, _ := json.Marshal(ctx.FormValues())
	apiLog := map[string]interface{}{
		"start_time": time.Now(),
		"user_name":  agent_name,
		"postData":   string(postData),
		"apiName":    "获取注单信息",
		"ip_info":    ctx.RemoteAddr(),
		"log_type":   "api",
	}

	//联调状态统计 只要有联调请求则统计数据
	is_debug := models.StatisticsLog{}.ApiStatistics(apiLog["apiName"].(string), agent_name)

	if agent_name == "" || token == "" {
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
	checkEequest := helper.CheckEequest(token, []string{securityKey, deadline, agent_name})

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

	if is_test == "" {
		//判断调用频率间隔时间
		has, errCode := this.getListFeed(agent_name)
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

	//根据代理商获取代理商ID
	agent_info, err := models.Agent{}.GetAgentInfo(agent_name)
	if err != nil || agent_info == nil {
		apiLog["code"] = 9004
		apiLog["text"] = config.ErrCode(9004)
		apiLog["result"] = ""
		apiLog["end_time"] = time.Now()
		go models.ApiLog{}.Insert(&apiLog)
		ctx.JSON(iris.Map{
			"code":   9004,
			"text":   config.ErrCode(9004), //代理不存在
			"result": " ",
		})
		return
	}

	//验证通过进行数据获取操作
	startDate, endDate := "", ""
	list_data, err := this.getOrderList(agent_info.Id, deadline, startDate, endDate)

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

	//fmt.Println(list_data)

	//联调账号联调成功次数统计
	if is_debug {
		models.StatisticsLog{}.ApiSucceds(apiLog["apiName"].(string), agent_name) //联调账号联调成功次数统计
	}

	apiLog["code"] = 0
	apiLog["text"] = config.ErrCode(0)
	apiLog["result"] = ""
	apiLog["end_time"] = time.Now()
	go models.ApiLog{}.Insert(&apiLog)
	ctx.JSON(iris.Map{
		"code" : 0,
		"text" : config.ErrCode(9999),
		"result" : iris.Map{
			"next_time" : list_data[len(list_data)-1].End_time.Format("2006-01-02 15:04:05"),
			"data" : list_data,
		},
	})

}

/**
 * @api {post} /getDateList 代理商根据时间段获取注单信息
 * @apiDescription 代理商根据时间段获取注单信息
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
					"round_no": "3c4c0e4df67432bd",
					"game_id": 93,
					"game_hall_id": 0,
					"game_name": "龙虎 ",
					"server_name": "15",
					"total_bet_score": 70,
					"valid_bet_score_total": 70,
					"game_result": "8;5",
					"total_win_score": 70,
					"user_name": "test123451",
					"is_mark": 1,
					"dwRound": "",
					"remark": "24;53",
					"_id": "5a056e10e1382363171f7ec3",
					"game_period": "248-56",//靴+局
					"start_time": "2017-11-10 17:15:06"
				}
			]
		},
		"text": "Success"
	}
*/
func (this UserCash) GetOrderListByDate(ctx context.Context) {
	//白名单验证
	if res := middleware.IpLimit(ctx); !res{
		return
	}

	agent_name := ctx.FormValue("agent")
	start_date := ctx.FormValue("start_date")
	end_date := ctx.FormValue("end_date")
	token := ctx.FormValue("token")
	is_test := ctx.FormValue("is_test") //压测使用
	//记录调用日志
	postData, _ := json.Marshal(ctx.FormValues())
	apiLog := map[string]interface{}{
		"start_time": time.Now(),
		"user_name":  agent_name,
		"postData":   string(postData),
		"apiName":    "时间段获取注单信息",
		"ip_info":    ctx.RemoteAddr(),
		"log_type":   "api",
	}

	//联调状态统计 只要有联调请求则统计数据
	is_debug := models.StatisticsLog{}.ApiStatistics(apiLog["apiName"].(string), agent_name)

	if agent_name == "" || start_date == "" || token == "" || end_date == "" {
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
	checkEequest := helper.CheckEequest(token, []string{securityKey, start_date, end_date, agent_name})

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

	// 时间验证
	has, err_code := this.checkStartAndEndDate(start_date, end_date)
	if !has {
		apiLog["code"] = err_code
		apiLog["text"] = config.ErrCode(err_code)
		apiLog["result"] = ""
		apiLog["end_time"] = time.Now()
		go models.ApiLog{}.Insert(&apiLog)
		ctx.JSON(iris.Map{
			"code":   err_code,
			"text":   config.ErrCode(err_code),
			"result": "",
		})
		return
	}

	// 时间跨度
	has, error_code := this.checkDatesSpacing(start_date, end_date)
	if !has {
		apiLog["code"] = error_code
		apiLog["text"] = config.ErrCode(error_code)
		apiLog["result"] = ""
		apiLog["end_time"] = time.Now()
		go models.ApiLog{}.Insert(&apiLog)
		ctx.JSON(iris.Map{
			"code":   error_code,
			"text":   config.ErrCode(error_code),
			"result": "",
		})
		return
	}

	if is_test == "" {
		//判断调用频率间隔时间
		has, errCode := this.getListFeed(agent_name)
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

	//根据代理商获取代理商ID
	agent_info, err := models.Agent{}.GetAgentInfo(agent_name)
	if err != nil || agent_info == nil {
		apiLog["code"] = 9004
		apiLog["text"] = config.ErrCode(9004)
		apiLog["result"] = ""
		apiLog["end_time"] = time.Now()
		go models.ApiLog{}.Insert(&apiLog)
		ctx.JSON(iris.Map{
			"code":   9004,
			"text":   config.ErrCode(9004), //代理不存在
			"result": " ",
		})
		return
	}

	//验证通过进行数据获取操作
	deadline := ""
	list_data, err := this.getOrderList(agent_info.Id, deadline, start_date, end_date)
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
		models.StatisticsLog{}.ApiSucceds(apiLog["apiName"].(string), agent_name) //联调账号联调成功次数统计
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

}

/**
 * 获取注单列表
 */
func (UserCash) getOrderList(agent_id int, deadline string, startDate string, endDate string) ([]models.UserChartInfo, error) {

	//获取单次最大获取的数据量条数，可配置
	take := config.Common().GET_ORDER_LIST_MAX

	//根据最大ID获取数据
	findWhere := bson.M{"agent_id": agent_id, "is_mark": 1}
	if len(deadline) > 0 {
		deadline_utc, _ := time.ParseInLocation("2006-01-02 15:04:05 ", deadline, time.Local)
		//findWhere = bson.M{ "agent_id": agent_id,  "is_mark":1 ,"end_time":bson.M{"$gte":deadline_utc} }
		findWhere = bson.M{"agent_id": agent_id, "is_mark": 1, "end_time": bson.M{"$gte": deadline_utc}}
	}

	//根据时间段获取数据
	if len(startDate) > 0 && len(endDate) > 0 {
		startDate_utc, _ := time.ParseInLocation("2006-01-02 15:04:05 ", startDate, time.Local)
		endDate_utc, _ := time.ParseInLocation("2006-01-02 15:04:05 ", endDate, time.Local)
		//findWhere = bson.M{"agent_id": agent_id, "is_mark":1 , "end_time":bson.M{"$gte":startDate_utc, "$lte":endDate_utc} }
		findWhere = bson.M{"agent_id": agent_id, "is_mark": 1, "end_time": bson.M{"$gte": startDate_utc, "$lte": endDate_utc}}
	}

	user_chart_Info, err := models.UserChartInfo{}.GetListData(findWhere, take, "end_time")

	//fmt.Println(user_chart_Info)
	if err != nil || len(user_chart_Info) == 0 {
		return nil, err
	}
	return user_chart_Info, nil
}

/**
 * 监测时间间隔
 */
func (UserCash) getListFeed(agent_name string) (bool, int) {
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
 *	 检查日期
 */
func (UserCash) checkStartAndEndDate(start_date string, end_date string) (bool, int) {
	if start_date == "" || end_date == "" {
		return false, 9002
	}
	start_time, _ := time.ParseInLocation("2006-01-02 15:04:05", start_date, time.Local)
	end_time, _ := time.ParseInLocation("2006-01-02 15:04:05", end_date, time.Local)
	if end_time.Before(start_time) {
		return false, 9002
	}
	return true, 0
}

/**
 * 时间跨度验证
 */
func (UserCash) checkDatesSpacing(start_date string, end_date string) (bool, int) {
	t1, _ := time.ParseInLocation("2006-01-02 15:04:05", start_date, time.Local)
	t2, _ := time.ParseInLocation("2006-01-02 15:04:05", end_date, time.Local)
	date := config.Common().MAX_TIME_SPAN
	if t2.Unix()-t1.Unix() > 3600*24*date {
		return false, 2001
	}
	return true, 0
}


/**
 * @api {post} /getBetList 获取注单详情信息
 * @apiDescription 获取注单详情信息（代理商为联调账号时可用）
 * @apiGroup User
 * @apiPermission
 * @apiVersion 1.0.0
 * @apiParam {String} agent 代理商用户名
 * @apiParam {String} token token:SHA1('securityKey|agent')
 * @apiSuccessExample {json} 成功返回格式
	{
		"code": 0,
		"result": {
			"data": [
				{
					"user_name": "Qyi23qaqaq",//玩家登陆名
					"game_id": 91,//游戏id
					"game_hall_id": 0,//游戏厅id
					"game_name": "百家乐",//游戏名称
					"bet_money": 30000,//下注金额
					"bet_money_valid": 30000,//有效下注金额
					"round_no": "f4079b5d14b5df68",//局ID
					"add_time": "2017-12-08 16:48:39"//下注时间
				}
			]
		},
		"text": "Success"
	}
*/
func (this UserCash) GetBetList(ctx context.Context)  {
	//白名单验证
	if res := middleware.IpLimit(ctx); !res{
		return
	}

	agent_name := ctx.FormValue("agent")
	token := ctx.FormValue("token")

	//记录调用日志
	postData, _ := json.Marshal(ctx.FormValues())
	apiLog := map[string]interface{}{
		"start_time": time.Now(),
		"user_name":  agent_name,
		"postData":   string(postData),
		"apiName":    "获取下注信息（联调）",
		"ip_info":    ctx.RemoteAddr(),
		"log_type":   "api",
	}

	if agent_name == "" || token == "" {
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

	//联调状态统计 只要有联调请求则统计数据
	is_debug := models.StatisticsLog{}.ApiStatistics(apiLog["apiName"].(string), agent_name)

	//不是联调账号，返回错误
	if ! is_debug {
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
	checkEequest := helper.CheckEequest(token, []string{securityKey, agent_name})

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

	//判断调用频率间隔时间
	has, errCode := this.getListFeed(agent_name)
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

	//根据代理商获取代理商ID
	agent_info, err := models.Agent{}.GetAgentInfo(agent_name)
	if err != nil || agent_info == nil {
		apiLog["code"] = 9004
		apiLog["text"] = config.ErrCode(9004)
		apiLog["result"] = ""
		apiLog["end_time"] = time.Now()
		go models.ApiLog{}.Insert(&apiLog)
		ctx.JSON(iris.Map{
			"code":   9004,
			"text":   config.ErrCode(9004), //代理不存在
			"result": " ",
		})
		return
	}
	res, err := models.UserOrder{}.GetListData(bson.M{"agent_id": agent_info.Id}, 30, "-add_time")
	if err != nil || res == nil || len(res) == 0{
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
		models.StatisticsLog{}.ApiSucceds(apiLog["apiName"].(string), agent_name) //联调账号联调成功次数统计
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
			"data": res,
		},
	})
}