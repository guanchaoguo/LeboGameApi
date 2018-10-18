package middleware

import (
	"encoding/json"
	"github.com/garyburd/redigo/redis"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
	"github.com/tidwall/gjson"
	"golang-LeboGameApi/app/helper"
	"golang-LeboGameApi/app/models"
	"golang-LeboGameApi/config"
	"golang-LeboGameApi/lib"
	"strings"
	"time"
)


/**
* IP限制
* @return mixed
* 一般来说代理商的IP就是厅主的IP，如果代理商和厅主的IP不一样，
* 就必须要在系统的白名单中把该代理商的IP添加入所属的厅主白名单中
 */
func IpLimit(ctx context.Context) bool {
	agent_name := ctx.FormValue("agent")
	if agent_name == "" {
		ctx.JSON(iris.Map{
			"code":   9002, //请求参数错误
			"text":   config.ErrCode(9002),
			"result": "",
		})
		return false
	}

	//获取厅主白名单redis
	redis_key := config.Common().WHITELIST
	c := helper.RedisClientDefault.Get()
	ips, _ := redis.String(c.Do("hget", redis_key, agent_name))
	defer c.Close()

	//重新获取数据白名单存入redis缓存
	if ips == "" || ips == "null" {
		agent_info, err := models.Agent{}.GetAgentInfo(agent_name)
		if err != nil || agent_info == nil {
			ctx.JSON(iris.Map{
				"code":   9004,
				"text":   config.ErrCode(9004), //代理不存在
				"result": "",
			})
			return false
		}

		white_list, err := models.WhiteList{}.GetWhiteList(agent_info.Parent_id)
		if err != nil || white_list == nil {
			ctx.JSON(iris.Map{
				"code":   9003, //请求参数错误
				"text":   config.ErrCode(9003),
				"result": "",
			})
			return false
		}

		//获取厅主
		hallAgent, _ := models.Agent{}.GetHallInfo(agent_info.Parent_id)
		//redis缓存白名单信息
		white_list.Agent_code = agent_info.Agent_code
		white_list.Account_type = agent_info.AccountType
		white_list.Agent_id2 = agent_info.Id
		white_list.Connect_mode = hallAgent.Connect_mode
		white_list_json, _ := json.Marshal(white_list)
		ips = string(white_list_json)
		redis.String(c.Do("hset", redis_key, agent_name, ips))
		//defer c.Close()
	}

	// 校验白名单
	ip_info := gjson.Get(ips, "ip_info").String()
	if ip_info == "" || (ip_info != "*" && !strings.Contains(ip_info, ctx.RemoteAddr())) {

		ctx.JSON(iris.Map{
			"code":   9003, //IP 限制
			"text":   config.ErrCode(9003),
			"result": "",
		})
		return false
	}
	return  true
}

/**
* 判断系统是否在维护当中
* @param string $username
* @param string $agent
* @return bool|void
*
 */
func CheckIsMaintain(agent_name string, user_name string) (bool, int) {

	if agent_name == "" {
		return false, 9002 //请求参数错误
	}

	// 存在用户则作管理身份校验
	if user_name != "" {
		//获取代理商code
		redis_key := config.Common().WHITELIST
		c := helper.RedisClientDefault.Get()
		defer c.Close()
		redis_data, err := redis.String(c.Do("hget", redis_key, agent_name))
		agent_code := gjson.Get(redis_data, "agent_code").String()

		decry_username := lib.Crypto{}.Decrypt(agent_code + user_name)
		user_info, err := models.User{}.GetInfo(decry_username)
		if err != nil {
			return false, 1007 // 用户不存在
		}

		// 管理员全天登录
		if user_info != nil && user_info.User_rank == 2 {
			return true, 9999 //   成功  暂时没有维护
		}
	}

	// 获取系统维护信息 没有维护信息则不维护
	// systemMintain,err := models.SystemMintain{}.GetInfo()

	redis_name := config.Common().GAME_MT_ING
	c := helper.RedisClientDefault.Get()
	defer c.Close()

	maintain, _ := redis.String(c.Do("get", redis_name))
	if maintain == "" {
		return true, 9999 //   成功  暂时没有维护
	}

	start_date := gjson.Get(maintain, "start_date").String()
	end_date := gjson.Get(maintain, "end_date").String()

	//判断当前时间是否在系统维护时间内
	now_time := time.Unix(time.Now().Unix(), 0)
	start_time, _ := time.ParseInLocation("2006-01-02 15:04:05", start_date, time.Local)
	end_time, _ := time.ParseInLocation("2006-01-02 15:04:05", end_date, time.Local)

	if start_time.Before(now_time) && end_time.After(now_time) {
		return false, 9001 // 系统维护
	}
	return true, 9999 //    成功  暂时没有维护
}
