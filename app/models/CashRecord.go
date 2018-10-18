package models

import (
	"gopkg.in/mgo.v2/bson"
)

/**
现金流记录model
*/
import (
	"fmt"
	"time"
)

type CashRecord struct {
	Order_sn      string    `json:"order_sn"`
	Uid           int       `json:"uid"`
	User_name     string    `json:"user_name"`
	Agent_id      int       `json:"agent_id"`
	Hall_id       int       `json:"hall_id"`
	Type          int       `json:"type"`
	Amount        float64   `json:"amount"`
	Status        int       `json:"status"`
	User_money    float64   `json:"user_money"`
	Desc          string    `json:"desc"`
	Admin_user    string    `json:"admin_user"`
	Admin_user_id int       `json:"admin_user_id"`
	Cash_no       string    `json:"cash_no"`
	Add_time      time.Time `json:"add_time"`
	Pkey          string    `json:"pkey"`
}

/**
 * 现金流记录
 * @param data map[string]interface{} 插入的数据
 * @return     bool 返回结果
 */
func (CashRecord) Insert(data *map[string]interface{}) bool {
	session := GetMongodb()
	if session == nil {
		return false
	}
	defer session.Close()
	c := session.DB(mongodbName).C("cash_record")
	err := c.Insert(data)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

func (CashRecord) Remove(order_sn string) bool {
	session := GetMongodb()
	if session == nil {
		return false
	}
	defer session.Close()
	c := session.DB(mongodbName).C("cash_record")
	err := c.Remove(&map[string]string{"order_sn":order_sn})
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}
/**
 * 通过pkey获取数据
 * @param pkey string pkey
 * @return    *CashRecord 结果
 * @return    error 错误值
 */
func (CashRecord) GetInfoByPkey(pkey string) (*CashRecord, error) {
	session := GetMongodb()
	defer session.Close()
	c := session.DB(mongodbName).C("cash_record")
	result := CashRecord{}
	where := bson.M{"pkey": pkey}
	// 查询是否存在该厅住的联调信息
	err := c.Find(&where).One(&result)
	return &result, err

}

/**
 * 测试
 */
func (CashRecord) Test() {
	session := GetMongodb()
	defer session.Close()
	c := session.DB(mongodbName).C("cash_record")
	c.Find("")
}
