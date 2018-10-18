package models

/**
api调用日志
*/
import (
	"fmt"
	"time"
)

type ApiLog struct {
	Agent      string    `json:"agent"`
	Api_url    string    `json:"api_url"`
	Otype_name string    `json:"otype_name"`
	Ip_info    string    `json:"ip_info"`
	Start_time time.Time `json:"start_time"`
	End_time   time.Time `json:"end_time"`
}

/**
 * 插入日志
 * @param data map 要出入的数据
 * @return     bool 返回结果
 */
func (ApiLog) Insert(data *map[string]interface{}) bool {
	session := GetMongodb()
	if session == nil {
		return false
	}
	defer session.Close()

	c := session.DB(mongodbName).C("api_log")
	err := c.Insert(data)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

/**
 * 登录日志
 * @param data map 要出入的数据
 * @return     bool 返回结果
 */
func (ApiLog) LoginLog(data *map[string]interface{}) bool {
	session := GetMongodb()
	if session == nil {
		return false
	}
	defer session.Close()
	c := session.DB(mongodbName).C("api_login_log")
	err := c.Insert(data)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}
