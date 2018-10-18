package controllers

import (
	"encoding/json"
	"github.com/garyburd/redigo/redis"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
	"github.com/tidwall/gjson"
	"golang-LeboGameApi/app/helper"
	"golang-LeboGameApi/app/models"
	"golang-LeboGameApi/config"
	"strconv"
	"time"
	"golang-LeboGameApi/app/middleware"
)

const GO_TIME = "2006-01-02 15:04:05"

type Key struct{}

/**
 * @api {post} /getKey 接入商获取SecurityKey
 * @apiGroup User
 * @apiPermission
 * @apiVersion 1.0.0
 * @apiParam {String} agent 代理商用户名
 * @apiParam {String} token token:SHA1('agent')
 * @apiSuccessExample {json} 成功返回格式
	{
		"code": 0,
		"result": {
			"expiration": "2017-12-22 13:16:22",
			"security_key": "25bc4e45a213f80eb56247b5e009a4bd1c19f9ff"
		},
		"text": "Success"
	}
*/
func (Key) Index(ctx context.Context) {

	//白名单验证
	if res := middleware.IpLimit(ctx); !res{
		return
	}

	agent_name := ctx.FormValue("agent")
	token := ctx.FormValue("token")

	if agent_name == "" || token == "" {
		ctx.JSON(iris.Map{
			"code":   9002,
			"text":   config.ErrCode(9002),
			"result": "",
		})
		return
	}

	jsonPost, _ := json.Marshal(ctx.FormValues())
	api_name := "获取SecurityKey"
	api_log := make(map[string]interface{})
	api_log["user_name"] = agent_name
	api_log["start_time"] = time.Now()
	api_log["postData"] = string(jsonPost)
	api_log["apiName"] = api_name
	api_log["ip_info"] = ctx.RemoteAddr()
	api_log["log_type"] = "api"

	//联调状态统计 只要有联调请求则统计数据
	var debugging bool = false
	res := models.StatisticsLog{}.ApiStatistics(api_name, agent_name)
	if res {
		debugging = true
	}

	//token认证对比
	checkEequest := helper.CheckEequest(token, []string{agent_name})
	if !checkEequest {
		ctx.JSON(iris.Map{
			"code":   9002,
			"text":   config.ErrCode(9002),
			"result": "",
		})
		return
	}

	var security_key string
	var expiration string
	//获取代理商白名单信息
	redis_key := config.Common().WHITELIST
	c := helper.RedisClientDefault.Get()
	whiteList, _ := redis.String(c.Do("hget", redis_key, agent_name))
	defer c.Close()
	//没有缓存数据，读取数据库数据（由于在白名单ip中间件的时候已经判断了，如果在这里没有缓存就直接返回ip限制）
	if whiteList == "" {
		ctx.JSON(iris.Map{
			"code":   9003, //白名单不存在，返回ip限制提示
			"text":   config.ErrCode(9003),
			"result": "",
		})
		return
		////获取代理商信息
		//agent_info,err := models.Agent{}.GetAgentInfo(agent_name)
		//if err != nil || agent_info == nil{
		//	ctx.JSON( iris.Map{
		//		"code" : 9004,
		//		"text" :config.ErrCode(9004) ,//代理不存在
		//		"result": " ",
		//	})
		//	return
		//}
		////根据厅主获取白名单信息
		//white_list ,err:= models.WhiteList{}.GetWhiteList(agent_info.Parent_id)
		//if err != nil || white_list == nil {
		//	ctx.JSON(iris.Map{
		//		"code":   9003, //白名单不存在，返回ip限制提示
		//		"text":   config.ErrCode(9003),
		//		"result": "",
		//	})
		//	return
		//}
		//now_time := time.Unix(time.Now().Unix(), 0)
		//start_time, _ := time.ParseInLocation(GO_TIME, white_list.Seckey_exp_date, time.Local)
		////时间过期，重新生成key
		//if now_time.After(start_time){
		//	_,has :=updateWhitelist(agent_name,agent_info.Parent_id)
		//	if !has {
		//		api_log["code"] = 9006
		//		api_log["text"] = config.ErrCode(9006)
		//		api_log["result"] = ""
		//		api_log["end_time"] = time.Now()
		//		go models.ApiLog{}.Insert(api_log)
		//
		//		ctx.JSON(iris.Map{
		//			"code" : 9006,
		//			"text" : config.ErrCode(9006),
		//			"result" : "",
		//		})
		//		return
		//	}
		//	// 再次获取白名单信息
		//	white_list ,_= models.WhiteList{}.GetWhiteList(agent_info.Parent_id)
		//}
		//
		////redis缓存白名单信息
		//white_list.Agent_code = agent_info.Agent_code
		//white_list.Account_type = agent_info.AccountType
		//white_list.Agent_id2 = agent_info.Id
		//white_list_json ,_:= json.Marshal(white_list)
		//key_name := config.Common().WHITELIST
		//c := helper.GetRedis()
		//redis.String(c.Do("hset",key_name, agent_name, string(white_list_json)))
		//defer c.Close()
		//
		//security_key = white_list.Agent_seckey
		//expiration = white_list.Seckey_exp_date
	} else {
		//判断key的有效时间过不过期
		seckey_exp_date := gjson.Get(whiteList, "seckey_exp_date").String()
		seckey_exp_date1, _ := time.ParseInLocation(GO_TIME, seckey_exp_date, time.Local)

		security_key = gjson.Get(whiteList, "agent_seckey").String()
		expiration = seckey_exp_date
		//时间过期，重新生成key
		if time.Now().Unix() > seckey_exp_date1.Unix() {
			hall_id, _ := strconv.Atoi(gjson.Get(whiteList, "agent_id").String())
			agent_id, _ := strconv.Atoi(gjson.Get(whiteList, "agent_id2").String())
			connect_mode := gjson.Get(whiteList, "connect_mode").Int()
			account_type, _ := strconv.Atoi(gjson.Get(whiteList, "account_type").String())
			agent_code := gjson.Get(whiteList, "agent_code").String()
			_, has := updateWhitelist(agent_name, hall_id)
			if !has {
				api_log["code"] = 9006
				api_log["text"] = config.ErrCode(9006)
				api_log["result"] = ""
				api_log["end_time"] = time.Now()
				go models.ApiLog{}.Insert(&api_log)

				ctx.JSON(iris.Map{
					"code":   9006,
					"text":   config.ErrCode(9006),
					"result": "",
				})
				return
			}
			// 再次获取白名单信息
			white_list, _ := models.WhiteList{}.GetWhiteList(hall_id)

			//redis缓存白名单信息
			white_list.Agent_code = agent_code
			white_list.Account_type = account_type
			white_list.Agent_id2 = agent_id
			white_list.Connect_mode = connect_mode
			white_list_json, _ := json.Marshal(white_list)
			key_name := config.Common().WHITELIST
			c := helper.RedisClientDefault.Get()
			redis.String(c.Do("hset", key_name, agent_name, string(white_list_json)))
			defer c.Close()

			security_key = white_list.Agent_seckey
			expiration = white_list.Seckey_exp_date
		}

	}

	api_log["code"] = 0
	res_json, _ := json.Marshal(iris.Map{
		"security_key": security_key,
		"expiration":   expiration,
	})
	api_log["text"] = config.ErrCode(9999)
	api_log["result"] = string(res_json)
	api_log["end_time"] = time.Now()
	go models.ApiLog{}.Insert(&api_log)

	//联调账号联调成功次数统计
	if debugging {
		models.StatisticsLog{}.ApiSucceds(api_name, agent_name) //联调账号联调成功次数统计
	}

	ctx.JSON(iris.Map{
		"code": 0,
		"text": config.ErrCode(9999),
		"result": iris.Map{
			"security_key": security_key,
			"expiration":   expiration,
		},
	})

}

/*
* 更新代理商白名单
* SecurityKey生成规则：根据代理商用户名和mt_rand(10,100000)随机数拼接，
* 然后字符串打散处理，最后进行 sha1()加密返回
* SecurityKey的有效时间为一个月
 */
func updateWhitelist(agent_name string, hall_id int) (map[string]string, bool) {

	str := agent_name + string(helper.RandInt64(10, 100000))
	str = helper.Shuffle_str(str)
	securityKey := helper.SHA1(str)
	seckey_exp_date := time.Now().AddDate(0, 0, config.Common().KEY_MAX_VALID_TIME).Format("2006-01-02 15:04:05")

	res, err := models.WhiteList{}.UpdateKey(hall_id, seckey_exp_date, securityKey)
	result := map[string]string{
		"securityKey":     securityKey,
		"seckey_exp_date": seckey_exp_date,
	}
	if err != nil || !res {
		return nil, false
	}
	return result, true

}

/**
获取厅主SecurityKey
*/
func GetSecurityKey(agent_name string) (interface{}, bool) {

	if agent_name == "" {
		return 9004, false
	}

	//获取厅主白名单redis
	redis_key := config.Common().WHITELIST
	c := helper.RedisClientDefault.Get()
	redis_data, err := redis.String(c.Do("hget", redis_key, agent_name))
	defer c.Close()
	//没有redis缓存，重新查数据库
	if err != nil || redis_data == "" {
		agentInfo, _ := models.Agent{}.GetAgentInfo(agent_name)
		whitelist, _ := models.WhiteList{}.GetWhiteList(agentInfo.Parent_id)
		if whitelist == nil {
			return 9003, false
		}
		//获取厅主
		hallAgent, _ := models.Agent{}.GetHallInfo(agentInfo.Parent_id)

		whitelist.Agent_code = agentInfo.Agent_code
		whitelist.Account_type = agentInfo.AccountType
		whitelist.Agent_id2 = agentInfo.Id
		whitelist.Connect_mode = hallAgent.Connect_mode
		whitelist_json, _ := json.Marshal(whitelist)
		redis.String(c.Do("hset", redis_key, agent_name, string(whitelist_json)))

		//判断key的有效时间过不过期
		seckey_exp_date, _ := time.ParseInLocation(GO_TIME, whitelist.Seckey_exp_date, time.Local)
		if time.Now().Unix() > seckey_exp_date.Unix() {
			return 9007, false
		}
		return whitelist.Agent_seckey, true
	}

	//判断key的有效时间过不过期
	seckey_exp_date := gjson.Get(redis_data, "seckey_exp_date").String()
	seckey_exp_date1, _ := time.ParseInLocation(GO_TIME, seckey_exp_date, time.Local)
	if time.Now().Unix() > seckey_exp_date1.Unix() {
		return 9007, false
	}
	return gjson.Get(redis_data, "agent_seckey").String(), true

}
