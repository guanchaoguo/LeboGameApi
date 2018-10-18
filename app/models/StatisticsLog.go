package models

import (
	"github.com/garyburd/redigo/redis"
	"github.com/tidwall/gjson"
	"golang-LeboGameApi/app/helper"
	"golang-LeboGameApi/config"
	"gopkg.in/mgo.v2/bson"
)

/**
联调账号日志
*/
type StatisticsLog struct {
	ApiName string `json:"api_name"`
	Agent   string `json:"agent"`
	Status  int    `json:"status"`
	Succeds int    `json:"succeds"`
	Sum     int    `json:"sum"`
}

/**
 * 联调接口调用总次数统计
 * @param apiName string api名称
 * @param agentName string 代理商名称
 * @return
 */
func (StatisticsLog) ApiStatistics(apiName string, agentName string) bool {
	//查询是否为联调账号
	/*is_debug := Agent{}.IsDebugAccount(agentName)
	if  ! is_debug {
		return false
	}*/
	redis_key := config.Common().WHITELIST
	c := helper.GetRedis("default")
	defer c.Close()
	redis_data, _ := redis.String(c.Do("hget", redis_key, agentName))
	account_type := gjson.Get(redis_data, "account_type").Int()
	if account_type != 3 {
		return false
	}
	session := GetMongodb()
	defer session.Close()
	//判断该联调代理数据是否存在 否则更新操作
	db := session.DB(mongodbName).C("api_statistics_log")

	result := StatisticsLog{}
	where := bson.M{"apiName": apiName, "agent": agentName}
	// 查询是否存在该厅住的联调信息
	db.Find(&where).One(&result)
	if result.Agent == "" {
		// 第一次统计数据
		insert_data := map[string]interface{}{
			"apiName": apiName,
			"agent":   agentName,
			"status":  0,
			"succeds": 0,
			"sum":     1,
		}
		db.Insert(insert_data)
	} else {
		// 更新次数
		update_data := bson.M{"$inc": bson.M{"sum": 1}}
		db.Update(&where, update_data)
	}

	return true
}

/**
 * 联调接口调用成功次数统计
 * @param apiName string api名称
 * @param agentName string 代理商名称
 * @return
 */
func (StatisticsLog) ApiSucceds(apiName string, agentName string) bool {
	//查询是否为联调账号
	/*is_debug := Agent{}.IsDebugAccount(agentName)
	if  ! is_debug {
		return false
	}*/
	redis_key := config.Common().WHITELIST
	c := helper.GetRedis("default")
	defer c.Close()
	redis_data, _ := redis.String(c.Do("hget", redis_key, agentName))
	account_type := gjson.Get(redis_data, "account_type").Int()
	if account_type != 3 {
		return false
	}

	session := GetMongodb()
	defer session.Close()
	//修改联调账号成功次数
	db := session.DB(mongodbName).C("api_statistics_log")

	where := bson.M{"apiName": apiName, "agent": agentName}
	update_data := bson.M{"$inc": bson.M{"succeds": 1}, "$set": bson.M{"status": 1}}
	db.Update(where, update_data)
	return true
}
