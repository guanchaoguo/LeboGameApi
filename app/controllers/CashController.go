package controllers

/**
 * 现金控制器
 */

import (
	"encoding/json"
	"github.com/garyburd/redigo/redis"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
	"golang-LeboGameApi/app/helper"
	"golang-LeboGameApi/app/middleware"
	"golang-LeboGameApi/app/models"
	"golang-LeboGameApi/config"
	"golang-LeboGameApi/lib"
	"strconv"
	"time"
	//"github.com/astaxie/beego/session"
	"github.com/tidwall/gjson"
	"fmt"
)

type CashController struct {
}

func (CashController) Test(ctx context.Context) {
	ctx.JSON(iris.Map{
		"code":   0,
		"text":   config.ErrCode(9999),
		"result": "",
	})
}

/**
 * @api {post} /deposit 会员充值
 * @apiDescription 会员将接入商钱包的钱，部分或者全部提取到游戏供应商的账户上进行游戏
 * @apiGroup User
 * @apiPermission
 * @apiVersion 1.0.0
 * @apiParam {String} agent 代理商用户名
 * @apiParam {String} username 用户登录名称
 * @apiParam {float} amount 充值金额
 * @apiParam {String} token token:SHA1('securityKey|username|amount|agent')
 * @apiSuccessExample {json} 成功返回格式
	{
		"code": 0,
		"result": {
			"amount": 100,//充值金额
			"order_sn": "LAA31303521259023"//订单号
		},
		"text": "Success"
	}
*/
func (CashController) Deposit(ctx context.Context) {
	//白名单验证
	if res := middleware.IpLimit(ctx); !res{
		return
	}

	agent_name := ctx.FormValue("agent")
	user_name := ctx.FormValue("username")
	amount := ctx.FormValue("amount")
	//deposit_type,_ := strconv.Atoi(ctx.FormValue("deposit_type"))//会员取消派彩异常充值
	deposit_type := ctx.FormValue("deposit_type") //会员取消派彩异常充值
	is_test := ctx.FormValue("is_test")           //压测使用

	token := ctx.FormValue("token")

	var cash_type int //充值类型
	postData, _ := json.Marshal(ctx.FormValues())
	//记录调用日志
	apiLog := map[string]interface{}{
		"start_time": time.Now(),
		"user_name":  agent_name,
		"postData":   string(postData),
		"apiName":    "会员充值",
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

	//存款类型 1：正常存款，2：会员取消派彩异常充值
	if deposit_type != "2" {
		deposit_type = "1"
		cash_type = 1 //转账
	} else {
		cash_type = 10 //取消派彩变更
		apiLog["apiName"] = "会员取消派彩异常充值"
	}

	//参数验证
	float_amount, _ := strconv.ParseFloat(amount, 64)

	if agent_name == "" || user_name == "" || token == "" || float_amount >= (100000000-1) || float_amount <= 0 {

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
	//token验证
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
	param := []string{securityKey, user_name, amount, agent_name}
	if deposit_type == "2" {
		param = append(param, deposit_type)
	}
	checkEequest := helper.CheckEequest(token, param)

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
	if c == nil {
		apiLog["code"] = 9004
		apiLog["text"] = config.ErrCode(9004)
		apiLog["result"] = "default redis error"
		apiLog["end_time"] = time.Now()
		go models.ApiLog{}.LoginLog(&apiLog)
		fmt.Println("default redis error")
		ctx.JSON(iris.Map{
			"code":   9004,
			"text":   config.ErrCode(9004),
			"result": "",
		})
		return
	}

	redis_data, _ := redis.String(c.Do("hget", redis_key, agent_name))
	agent_code := gjson.Get(redis_data, "agent_code").String()
	connect_mode := gjson.Get(redis_data, "connect_mode").Int()
	//如果是共享钱包返回代理错误
	if connect_mode == 1 {
		apiLog["code"] = 9004
		apiLog["text"] = config.ErrCode(9004)
		apiLog["result"] = "该代理的厅主是共享钱包模式，不能使用会员充值接口"
		apiLog["end_time"] = time.Now()
		go models.ApiLog{}.LoginLog(&apiLog)
		ctx.JSON(iris.Map{
			"code":   9004,
			"text":   config.ErrCode(9004),
			"result": "",
		})
		return
	}

	//会员名称加密
	decry_user_name := lib.Crypto{}.Decrypt(agent_code + user_name)
	lib.Crypto{}.Decrypt(agent_code + user_name)
	userInfo, err := models.User{}.GetUser(decry_user_name, agent_name)

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

	if userInfo == nil {

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

	if  userInfo.Account_state != 1{
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
	//充值
	//sql := "update `lb_user` set money=money+?, grand_total_money=grand_total_money+? where uid=?"
	//res:= models.User{}.Query(sql,float_amount,userInfo.Uid)
	data := map[string]interface{}{
		"money":             userInfo.Money + float_amount,
		"grand_total_money": userInfo.Money + float_amount,
	}

	//fmt.Println(data)
	//res := models.User{}.Update(userInfo.Uid, data)
	res, session := models.User{}.UpdateMoney(userInfo.Uid, data)
	defer session.Close()
	//充值失败
	if !res {
		session.Rollback()
		apiLog["code"] = 3003
		apiLog["text"] = config.ErrCode(3003)
		apiLog["result"] = "mysql:用户充值失败"
		apiLog["end_time"] = time.Now()
		go models.ApiLog{}.Insert(&apiLog)

		ctx.JSON(iris.Map{
			"code":   3003,
			"text":   config.ErrCode(3003),
			"result": "",
		})
		return
	}

	order_sn := helper.CreateOrderSn("L")
	//userInfo2, _ := models.User{}.GetUserByID(userInfo.Uid)
	//现金记录
	insert_data := map[string]interface{}{
		"order_sn":      order_sn,
		"uid":           userInfo.Uid,
		"user_name":     lib.Crypto{}.Encrypt(userInfo.User_name),
		"type":          cash_type,
		"amount":        float_amount,
		"status":        3,
		"user_money":    data["money"],
		"desc":          "流水号" + order_sn,
		"admin_user":    "'system-api",
		"admin_user_id": 0,
		"cash_no":       order_sn,
		"agent_id":      userInfo.Agent_id,
		"hall_id":       userInfo.Hall_id,
		"pkey":          helper.MD5(agent_name + order_sn + config.Common().GAME_API_SUF),
		"add_time":      time.Now(),
	}
	//fmt.Println(insert_data)*/
	insert_res := models.CashRecord{}.Insert(&insert_data)

	if !insert_res {
		session.Rollback()
		//失败，回滚金额
		//data := map[string]interface{}{
		//	"money":             userInfo2.Money - float_amount,
		//	"grand_total_money": userInfo2.Money - float_amount,
		//}
		//models.User{}.Update(userInfo2.Uid, data)
		fmt.Println("mongodb:用户充值现金流记录失败，用户充值金额回滚")
		apiLog["code"] = 3003
		apiLog["text"] = config.ErrCode(3003)
		apiLog["result"] = "mongodb:用户充值现金流记录失败，用户充值金额回滚"
		apiLog["end_time"] = time.Now()
		go models.ApiLog{}.Insert(&apiLog)

		ctx.JSON(iris.Map{
			"code":   3003,
			"text":   config.ErrCode(3003),
			"result": "",
		})
		return
	}

	//调用日志记录
	result, _ := json.Marshal(iris.Map{
		"order_sn": order_sn,
		"amount":   amount,
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

	if is_test == "" {
		//统计充值记录
		models.StatisCashAgent{}.TotalScoreRecord(agent_name, redis_data, float_amount)
	}

	//用户增加充值、扣款后累计清除下注次数
	re := clearBetcount(userInfo.Uid)
	if ! re {
		session.Rollback()
		//失败，回滚金额
		/*data := map[string]interface{}{
			"money":             userInfo2.Money - float_amount,
			"grand_total_money": userInfo2.Money - float_amount,
		}
		models.User{}.Update(userInfo2.Uid, data)*/
		fmt.Println("redis:清除下注次数失败，用户充值金额回滚，现金流记录回滚order_sn:" + order_sn)
		//回滚现金流
		models.CashRecord{}.Remove(order_sn)

		apiLog["code"] = 3003
		apiLog["text"] = config.ErrCode(3003)
		apiLog["result"] = "redis:清除下注次数失败，用户充值金额回滚，现金流记录回滚order_sn:" + order_sn
		apiLog["end_time"] = time.Now()
		go models.ApiLog{}.Insert(&apiLog)

		ctx.JSON(iris.Map{
			"code":   3003,
			"text":   config.ErrCode(3003),
			"result": "",
		})
		return
	}

	//成功返回
	ctx.JSON(iris.Map{
		"code": 0,
		"text": config.ErrCode(9999),
		"result": iris.Map{
			"order_sn": order_sn,
			"amount":   strconv.FormatFloat(float_amount,'f',2,64),
		}})
	session.Commit()
	fmt.Println("会员充值成功，所有事务提交")
}

//用户增加充值、扣款后累计清除下注次数
func clearBetcount(uid int) bool{
	c := helper.RedisClientMonitor.Get()
	if c == nil {
		return  false
	}
	redis.String(c.Do("SET", "betcount:"+strconv.Itoa(uid), 0))
	defer c.Close()
	return  true
}

/**
 * @api {post} /withDrawal 会员取款
 * @apiDescription 会员将接入商钱包的钱， 取款时需调用
 * @apiGroup User
 * @apiPermission
 * @apiVersion 1.0.0
 * @apiParam {String} agent 代理商用户名
 * @apiParam {String} username 用户登录名称
 * @apiParam {float} amount 扣款金额
 * @apiParam {String} token token:SHA1('securityKey|username|agent')
 * @apiSuccessExample {json} 成功返回格式
	{
		"code": 0,
		"result": {
			"amount": 100,//扣款金额
			"order_sn": "LAA31303521259023"//订单号
		},
		"text": "Success"
	}
*/
func (CashController) WithDrawal(ctx context.Context) {
	//白名单验证
	if res := middleware.IpLimit(ctx); !res{
		return
	}

	agent_name := ctx.FormValue("agent")
	user_name := ctx.FormValue("username")
	amount := ctx.FormValue("amount")
	if amount == "" {
		amount = "0"
	}
	token := ctx.FormValue("token")
	postData, _ := json.Marshal(ctx.FormValues())
	//记录调用日志
	apiLog := map[string]interface{}{
		"start_time": time.Now(),
		"user_name":  agent_name,
		"postData":   string(postData),
		"apiName":    "会员取款",
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

	//参数验证
	float_amount, _ := strconv.ParseFloat(amount, 64)

	if agent_name == "" || user_name == "" || token == "" || float_amount <= 0 {

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

	//token验证
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
	param := []string{securityKey, user_name, amount, agent_name}
	checkEequest := helper.CheckEequest(token, param)
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
	if c == nil {
		apiLog["code"] = 9004
		apiLog["text"] = config.ErrCode(9004)
		apiLog["result"] = "default redis error"
		apiLog["end_time"] = time.Now()
		go models.ApiLog{}.LoginLog(&apiLog)
		fmt.Println("default redis error")
		ctx.JSON(iris.Map{
			"code":   9004,
			"text":   config.ErrCode(9004),
			"result": "",
		})
		return
	}
	redis_data, err := redis.String(c.Do("hget", redis_key, agent_name))
	agent_code := gjson.Get(redis_data, "agent_code").String()
	connect_mode := gjson.Get(redis_data, "connect_mode").Int()
	//如果是共享钱包返回代理错误
	if connect_mode == 1 {
		apiLog["code"] = 9004
		apiLog["text"] = config.ErrCode(9004)
		apiLog["result"] = "该代理的厅主是共享钱包模式，不能使用会员取款接口"
		apiLog["end_time"] = time.Now()
		go models.ApiLog{}.LoginLog(&apiLog)
		ctx.JSON(iris.Map{
			"code":   9004,
			"text":   config.ErrCode(9004),
			"result": "",
		})
		return
	}
	//会员名称加密
	decry_user_name := lib.Crypto{}.Decrypt(agent_code + user_name)
	userInfo, err := models.User{}.GetUser(decry_user_name, agent_name)
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

	if userInfo == nil {
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

	if  userInfo.Account_state != 1{
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

	//判斷余额
	if (userInfo.Money - float_amount) < 0 {
		apiLog["code"] = 3001
		apiLog["text"] = config.ErrCode(3001)
		apiLog["result"] = ""
		apiLog["end_time"] = time.Now()
		go models.ApiLog{}.Insert(&apiLog)

		ctx.JSON(iris.Map{
			"code":   3001,
			"text":   config.ErrCode(3001),
			"result": "",
		})
		return
	}
	//取款操作
	//sql := "UPDATE  lb_user SET money=money-?, grand_total_money=grand_total_money-? WHERE uid=?"
	//res := models.User{}.Query(sql,float_amount,userInfo.Uid)

	data := map[string]interface{}{
		"money":             userInfo.Money - float_amount,
		"grand_total_money": userInfo.Money - float_amount,
	}
	//res := models.User{}.Update(userInfo.Uid, data)
	res, session := models.User{}.UpdateMoney(userInfo.Uid, data)
	defer session.Close()
	//取款失败
	if !res {
		session.Rollback()
		apiLog["code"] = 3002
		apiLog["text"] = config.ErrCode(3002)
		apiLog["result"] = "mysql:用户取款失败"
		apiLog["end_time"] = time.Now()
		go models.ApiLog{}.Insert(&apiLog)

		ctx.JSON(iris.Map{
			"code":   3002,
			"text":   config.ErrCode(3002),
			"result": "",
		})
		return
	}
	//现金记录
	//userInfo2, _ := models.User{}.GetUserByID(userInfo.Uid)
	order_sn := helper.CreateOrderSn("L")

	insert_data := map[string]interface{}{
		"order_sn":      order_sn,
		"uid":           userInfo.Uid,
		"user_name":     lib.Crypto{}.Encrypt(userInfo.User_name),
		"type":          1, //转账
		"amount":        float_amount,
		"status":        4, //取款
		"user_money":    data["money"],
		"desc":          "流水号" + order_sn,
		"admin_user":    "'system-api",
		"admin_user_id": 0,
		"cash_no":       order_sn,
		"agent_id":      userInfo.Agent_id,
		"hall_id":       userInfo.Hall_id,
		"pkey":          helper.MD5(agent_name + order_sn + config.Common().GAME_API_SUF),
		"add_time":      time.Now(),
	}
	insert_res := models.CashRecord{}.Insert(&insert_data)

	if !insert_res {
		session.Rollback()
		//失败，回滚金额
		//data := map[string]interface{}{
		//	"money":             userInfo2.Money + float_amount,
		//	"grand_total_money": userInfo2.Money + float_amount,
		//}
		//models.User{}.Update(userInfo2.Uid, data)
		fmt.Println("mongodb:用户取款现金流记录失败，用户取款金额回滚")
		apiLog["code"] = 3002
		apiLog["text"] = config.ErrCode(3002)
		apiLog["result"] = "mongodb:用户取款现金流记录失败，用户取款金额回滚"
		apiLog["end_time"] = time.Now()
		go models.ApiLog{}.Insert(&apiLog)

		ctx.JSON(iris.Map{
			"code":   3002,
			"text":   config.ErrCode(3002),
			"result": "",
		})
		return
	}

	//调用日志记录
	result, _ := json.Marshal(iris.Map{
		"order_sn": order_sn,
		"amount":   amount,
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

	//用户增加充值、扣款后累计清除下注次数
	re := clearBetcount(userInfo.Uid)
	if ! re {
		session.Rollback()
		//失败，回滚金额
		/*data := map[string]interface{}{
			"money":             userInfo2.Money + float_amount,
			"grand_total_money": userInfo2.Money + float_amount,
		}
		models.User{}.Update(userInfo2.Uid, data)*/
		fmt.Println("redis:清除下注次数失败，用户取款金额回滚，现金流记录回滚order_sn:" + order_sn)
		//回滚现金流
		models.CashRecord{}.Remove(order_sn)

		apiLog["code"] = 3002
		apiLog["text"] = config.ErrCode(3002)
		apiLog["result"] = "redis:清除下注次数失败，用户取款金额回滚，现金流记录回滚order_sn:" + order_sn
		apiLog["end_time"] = time.Now()
		go models.ApiLog{}.Insert(&apiLog)

		ctx.JSON(iris.Map{
			"code":   3002,
			"text":   config.ErrCode(3002),
			"result": "",
		})
		return
	}

	//成功返回
	ctx.JSON(iris.Map{
		"code": 0,
		"text": config.ErrCode(9999),
		"result": iris.Map{
			"order_sn": order_sn,
			"amount":   strconv.FormatFloat(float_amount,'f',2,64),
		}})
	//当所有事物都成功提交
	fmt.Println("会员取款成功，所有事务提交")
	session.Commit()
}
