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
	"regexp"
	"strconv"
	"strings"
	"time"
	"fmt"
)

/**
 * 会员登录、注册控制器
 */

type AuthController struct {
}

/**
 * @api {post} /authorization 会员登录&注册
 * @apiDescription 会员登录到供应商进行游戏。首先供应商会检测用户是否存在，
					如果不存在，供应商自动创建账号，如果存在则返回游戏地址，直接进入游戏大厅
 * @apiGroup User
 * @apiPermission
 * @apiVersion 1.0.0
 * @apiParam {String} agent 代理商用户名
 * @apiParam {String} username 用户登录名称
 * @apiParam {String} account_type 玩家类型：1：正式玩家，2：试玩玩家，默认为1
 * @apiParam {String} login_type 登录类型：1：pc，2:h5，默认为1
 * @apiParam {String} token token:SHA1('securityKey|username|agent|login_type')/SHA1('securityKey|account_type|agent|login_type')
 * @apiSuccessExample {json} 成功返回格式
	{
		"code": 0,
		"result": "https://pc.lggame.co/game.php?uid=652e4eef5764df2b7d123",
		"text": "Success"
	}
*/
func (AuthController) Login(ctx context.Context) {
	//白名单验证
	if res := middleware.IpLimit(ctx); !res{
		return
	}

	agent_name := ctx.FormValue("agent")
	user_name := ctx.FormValue("username")
	account_type := ctx.FormValue("account_type")

	is_test := ctx.FormValue("is_test") //压测使用

	if account_type == "" {
		account_type = "1"
	}

	login_type := ctx.FormValue("login_type")

	if login_type == "" {
		login_type = "1"
	}

	token := ctx.FormValue("token")

	postData, _ := json.Marshal(ctx.FormValues())
	//记录调用日志
	apiLog := map[string]interface{}{
		"start_time": time.Now(),
		"agent":      agent_name,
		"postData":   string(postData),
		"apiName":    "玩家登录游戏",
		"ip_info":    ctx.RemoteAddr(),
		"log_type":   "login",
		"user_name":  user_name,
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

	//检测login_type
	enable_login_type := false
	for _, v := range [2]string{"1", "2"} {
		if v == login_type {
			enable_login_type = true
		}
	}

	//检测account_type
	enable_account_type := true
	//检测用户名
	match_username := true

	switch account_type {
	case "1":
		match, _ := regexp.MatchString("^[a-zA-Z0-9_]{3,20}$", user_name)
		match_username = match
		break
	case "2":
		break
	default:
		enable_account_type = false
	}

	//参数验证
	if agent_name == "" || token == "" || !enable_login_type || !enable_account_type || !match_username {

		apiLog["code"] = 9002
		apiLog["text"] = config.ErrCode(9002)
		apiLog["result"] = ""
		apiLog["end_time"] = time.Now()
		go models.ApiLog{}.LoginLog(&apiLog)

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

	checktoken := []string{securityKey, user_name, agent_name, login_type}
	if account_type == "2" {
		checktoken = []string{securityKey, account_type, agent_name, login_type}
	}
	checkEequest := helper.CheckEequest(token, checktoken)

	if !checkEequest {
		apiLog["code"] = 9002
		apiLog["text"] = config.ErrCode(9002)
		apiLog["result"] = ""
		apiLog["end_time"] = time.Now()
		go models.ApiLog{}.LoginLog(&apiLog)

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

	agentRedis, err := redis.String(c.Do("hget", redis_key, agent_name))
	agentCode := gjson.Get(agentRedis, "agent_code").String()

	//会员名称加密
	decry_user_name := lib.Crypto{}.Decrypt(agentCode + user_name)
	decry_username_md := lib.Crypto{}.Decrypt(user_name)

	// 保存用户名
	apiLog["user_name"] = agentCode + user_name

	userInfo, err := models.User{}.GetUser(decry_user_name, agent_name)
	if err != nil {
		apiLog["code"] = 9002
		apiLog["text"] = config.ErrCode(9002)
		apiLog["result"] = ""
		apiLog["end_time"] = time.Now()
		go models.ApiLog{}.LoginLog(&apiLog)

		ctx.JSON(iris.Map{
			"code":   9002,
			"text":   config.ErrCode(9002),
			"result": "",
		})
		return
	}

	//如果是试玩账号，先验证测试代理商的白名单
	if account_type == "2" {

		status, res := checkAgentTestWhiteList(ctx.RemoteAddr(), "agent_test")

		if !status {
			ctx.JSON(res)
			return
		}
	}
	//用户不存在，注册
	if userInfo == nil || account_type == "2" {
		//不读数据库，读缓存
		/*agentInfo ,_:= models.Agent{}.GetAgentInfo(agent_name)
		//代理错误
		if agentInfo == nil{

			apiLog["code"] = 9004
			apiLog["text"] = config.ErrCode(9004)
			apiLog["result"] = ""
			apiLog["end_time"] = time.Now()
			go models.ApiLog{}.LoginLog(apiLog)

			ctx.JSON(iris.Map{
				"code" : 9004,
				"text" : config.ErrCode(9004),
				"result" : "",
			})
			return
		}*/

		accountType, _ := strconv.Atoi(gjson.Get(agentRedis, "account_type").String())
		agentId, _ := strconv.Atoi(gjson.Get(agentRedis, "agent_id2").String())
		hallId, _ := strconv.Atoi(gjson.Get(agentRedis, "agent_id").String())
		hallName := gjson.Get(agentRedis, "agent_name").String()
		//检查联调玩家数量是否超过限制
		if accountType == 3 {
			playerNum := config.Common().TEST_USER_COUNT
			if playerNum == 0 {
				playerNum = 10
			}
			userNum := models.User{}.AgentUserNums(agentId)

			if userNum > playerNum {
				ctx.JSON(iris.Map{
					"code":   4001,
					"text":   config.ErrCode(4001),
					"result": "",
				})
				return
			}

		} else if accountType == 2 {

			ctx.JSON(iris.Map{
				"code":   1005,
				"text":   config.ErrCode(1005),
				"result": "",
			})
			return
		}
		//密码盐值
		salt := helper.Randomkeys(20)
		pwd := helper.MakePassword(6)
		alias := "api会员"
		var userMoney float64
		//如果为试玩玩家则指定将玩家绑定到uid =2 的玩家代理上面
		if account_type == "2" {
			// 查询玩家试玩代理名称
			test_user, _ := models.User{}.GetUserByID(2)
			test_agent, _ := models.Agent{}.GetAgentInfo(test_user.Agent_name)
			userMoney = test_user.Money
			hallId = test_user.Hall_id
			hallName = test_user.Hall_name
			agentId = test_user.Agent_id
			agent_name = test_user.Agent_name
			accountType = test_agent.AccountType
			agentCode = test_agent.Agent_code
			alias = "test member"
			useRand := helper.MD5(helper.CreateOrderSn(string(time.Now().UnixNano())))[0:14]
			decry_user_name = lib.Crypto{}.Decrypt(test_agent.Agent_code + useRand)
			decry_username_md = lib.Crypto{}.Decrypt(useRand)
		}

		user_rank := 0
		if accountType != 1 {
			user_rank = 1
		}
		if agentId == 0 && hallId == 0 {
			apiLog["code"] = 9004
			apiLog["text"] = config.ErrCode(9004)
			apiLog["result"] = ""
			apiLog["end_time"] = time.Now()
			go models.ApiLog{}.LoginLog(&apiLog)

			ctx.JSON(iris.Map{
				"code":   9004,
				"text":   config.ErrCode(9004),
				"result": "",
			})
			return
		}
		//创建用户
		dbUser := models.User{
			User_name:     decry_user_name,
			Username_md:   decry_username_md,
			Password:      pwd,
			Password_md:   pwd,
			Alias:         alias,
			Create_time:   time.Now().Format("2006-01-02 15:04:05"),
			Add_date:      time.Now().Format("2006-01-02 15:04:05"),
			Add_ip:        ctx.RemoteAddr(),
			Ip_info:       ctx.RemoteAddr(),
			Hall_id:       hallId,
			Hall_name:     hallName,
			Agent_id:      agentId,
			Agent_name:    agent_name,
			Salt:          salt,
			User_rank:     user_rank,
			Account_state: 1,
			Money:         userMoney,
			On_line:       "N",
			Language:      "zh-cn",
			Last_time:     "0000-00-00 00:00:00",
		}
		userInfo, err = dbUser.Insert()

		if userInfo == nil || err != nil {
			apiLog["code"] = 1005
			apiLog["text"] = config.ErrCode(1005)
			apiLog["result"] = ""
			apiLog["end_time"] = time.Now()
			go models.ApiLog{}.LoginLog(&apiLog)

			ctx.JSON(iris.Map{
				"code":   1005,
				"text":   config.ErrCode(1005),
				"result": "",
			})
			return
		}

		if accountType == 1 {
			//更新代理商玩家数
			go models.Agent{}.SubUserNum(agentId)
			//更新厅主代理商玩家数
			go models.Agent{}.SubUserNum(hallId)
		}

	}

	//非正常账号不能登录
	if userInfo.Account_state != 1 && userInfo.Account_state != 4 {
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

	//将登录时间转时间戳
	theTime, _ := time.ParseInLocation("2006-01-02 15:04:05", userInfo.Last_time, time.Local)

	//限制频繁登录10秒内
	if time.Now().Unix()-theTime.Unix() < 10 {
		apiLog["code"] = 1004
		apiLog["text"] = config.ErrCode(1004)
		apiLog["result"] = ""
		apiLog["end_time"] = time.Now()
		go models.ApiLog{}.LoginLog(&apiLog)

		ctx.JSON(iris.Map{
			"code":   1004,
			"text":   config.ErrCode(1004),
			"result": "",
		})
		return
	}

	//将用户信息存到redis,游戏前端使用
	session_id := helper.MD5(userInfo.User_name)
	uid := session_id[0:21]

	userInfo.Username2 = userInfo.User_name
	userInfo.User_name = lib.Crypto{}.Encrypt(userInfo.User_name)
	userInfo.Username_md = lib.Crypto{}.Encrypt(userInfo.Username_md)
	userInfo.Time = time.Now().Format("2006-01-02 15:04:05")
	userInfo.Agent_code = agentCode

	c = helper.RedisClientAccount.Get()
	if c == nil {
		fmt.Println("会员的redis连接错误！保存redis用户信息失败")
		apiLog["code"] = 1005
		apiLog["text"] = config.ErrCode(1005)
		apiLog["result"] = ""
		apiLog["end_time"] = time.Now()
		go models.ApiLog{}.LoginLog(&apiLog)

		ctx.JSON(iris.Map{
			"code":   1005,
			"text":   config.ErrCode(1005),
			"result": "",
		})
		return
	}
	userInfo_json, _ := json.Marshal(userInfo)
	redis.String(c.Do("SET", uid, string(userInfo_json)))
	defer c.Close()

	if is_test == "" {
		//更新登录数据
		update_data := map[string]interface{}{
			"on_line":  "Y",
			"ip_info":  ctx.RemoteAddr(),
			"token_id": uid,
		}
		if userInfo.Account_state == 4 {
			update_data["account_state"] = 1
		} else {
			update_data["account_state"] = userInfo.Account_state
		}
		//更新玩家数据
		go models.User{}.Update(userInfo.Uid, update_data)
	}

	c = helper.RedisClientDefault.Get()
	defer c.Close()

	url := ""
	switch login_type {
	case "1":
		pc_host, _ := redis.String(c.Do("hget", "GAMEHOST:URL", 1))
		if pc_host == "" {
			pc_host = config.Common().GAME_HOST_PC
		} else {
			pc_host += "/game.php"
		}
		url = pc_host + "?uid=" + uid
		break
	case "2":
		h5_host, _ := redis.String(c.Do("hget", "GAMEHOST:URL", 2))
		if h5_host == "" {
			h5_host = config.Common().GAME_HOST
		}
		url = h5_host + "?uid=" + uid
		break
	default:
		pc_host, _ := redis.String(c.Do("hget", "GAMEHOST:URL", 1))
		if pc_host == "" {
			pc_host = config.Common().GAME_HOST_PC
		} else {
			pc_host += "/game.php"
		}
		url = pc_host + "?uid=" + uid
	}

	//记录调用日志
	result, _ := json.Marshal(iris.Map{
		"url": url,
	})
	apiLog["code"] = 0
	apiLog["text"] = config.ErrCode(9999)
	apiLog["result"] = string(result)
	apiLog["end_time"] = time.Now()
	go models.ApiLog{}.LoginLog(&apiLog)

	//联调账号联调成功次数统计
	if is_debug {
		go models.StatisticsLog{}.ApiSucceds(apiLog["apiName"].(string), agent_name)
	}

	ctx.JSON(iris.Map{
		"code":   0,
		"text":   config.ErrCode(9999),
		"result": url,
	})
	return
}

func checkAgentTestWhiteList(clientIp string, agentName string) (bool, map[string]interface{}) {

	//获取厅主白名单redis
	redis_key := config.Common().WHITELIST
	c := helper.RedisClientDefault.Get()
	agentTestWhitlist, _ := redis.String(c.Do("hget", redis_key, agentName))
	defer c.Close()

	if agentTestWhitlist == "" {
		agent_info, _ := models.Agent{}.GetAgentInfo(agentName)
		white_list, _ := models.WhiteList{}.GetWhiteList(agent_info.Parent_id)

		if white_list == nil {
			res := iris.Map{
				"code":   9003, //IP 限制
				"text":   config.ErrCode(9003),
				"result": "",
			}
			return false, res
		}
		//redis缓存白名单信息
		white_list.Agent_code = agent_info.Agent_code
		white_list.Account_type = agent_info.AccountType
		white_list.Agent_id2 = agent_info.Id
		white_list_json, _ := json.Marshal(white_list)
		agentTestWhitlist = string(white_list_json)
		redis.String(c.Do("hset", redis_key, agentName, agentTestWhitlist))
		defer c.Close()

	}

	// 校验白名单
	ip_info := gjson.Get(agentTestWhitlist, "ip_info").String()
	if ip_info == "" || (ip_info != "*" && strings.Contains(ip_info, clientIp)) {
		res := iris.Map{
			"code":   9003, //IP 限制
			"text":   config.ErrCode(9003),
			"result": "",
		}
		return false, res
	}
	return true, nil
}
