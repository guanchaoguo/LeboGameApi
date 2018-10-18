package models

import (
	"fmt"
	//"github.com/go-xorm/xorm"
	"github.com/go-xorm/xorm"
)

type User struct {
	Uid               int     `json:"uid" xorm:"pk autoincr"`
	User_name         string  `json:"user_name" xorm:"unique"` //'玩家在第三方平台账号，带前缀'
	Username_md       string  `json:"username_md"`
	Password          string  `json:"password"` //密码
	Password_md       string  `json:"password_md"`
	User_rank         int     `json:"user_rank"`                             //用户等级，1为测试玩家，0为正常账号,2为管理员测试账号
	Alias             string  `json:"alias"`                                 //玩家昵称
	Add_date          string  `json:"add_date"`                              //注册时间
	Create_time       string  `json:"create_time"`                           //创建时间
	Last_time         string  `json:"last_time" xorm:"DateTime"`             //最后一次登录时间
	Add_ip            string  `json:"add_ip"`                                //注册IP
	Ip_info           string  `json:"ip_info"`                               //登录IP
	On_line           string  `json:"on_line"`                               //是否在线
	Account_state     int     `json:"account_state" xorm:"int(3) default 1"` //账号状态,1为正常（启用）,2为暂停使用（冻结）,3为停用
	Hall_id           int     `json:"hall_id"`                               //厅主ID
	Agent_id          int     `json:"agent_id"`                              //代理商ID
	Hall_name         string  `json:"hall_name"`                             //厅主登录名
	Agent_name        string  `json:"agent_name"`                            //代理商名称
	Salt              string  `json:"salt"`                                  //盐值
	Money             float64 `json:"money"`                                 //当前用户余额
	Grand_total_money float64 `json:"grand_total_money"`                     //充值扣款累计余额
	Token_id          string  `json:"token_id"`                              //token_id
	Language          string  `json:"language" xorm:"language"`              //语言
	Username2         string  `json:"username2" xorm:"-"`                    //用户名2
	Time              string  `json:"time" xorm:"-"`                         //时间
	Agent_code        string  `json:"agent_code" xorm:"-"`
}

func (*User) TableName() string {
	return "lb_user"
}

func (User) GetInfo(user_name string) (*User, error) {
	var user User
	has, err := engineSlave.Where("user_name=?", user_name).Get(&user)
	if err != nil || !has {
		fmt.Println(err)
		return nil, err
	}
	return &user, nil
}

/**
根据代理商名称、玩家名称 获取玩家信息
*/
func (User) GetUser(user_name string, agent_name string) (*User, error) {

	var user User
	//Cols("uid", "money","on_line", "account_state")
	has, err := engineSlave.
		Where("user_name=? and agent_name=?", user_name, agent_name).
		Get(&user)
	//defer engine.Close()
	//报错
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	//数据不存在
	if !has {
		return nil, nil
	}
	return &user, nil
}

/**
 * 原生语句执行（有事物）
 * @param sql  string 要执行的sql语句
 * @return bool 返回结果
 */
func (User) Query(sql string, amount float64, uid int) bool {
	//session := engine.NewSession()
	//err := session.Begin()
	_, err := engine.Exec(sql, amount, amount, uid)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

func (User) UpdateMoney(uid int, data map[string]interface{}) (bool, *xorm.Session)  {
	session := engine.NewSession()
	session.Begin()
	res, err := session.Table("lb_user").Where("uid=?", uid).Update(data)
	if res == 0 || err != nil {
		fmt.Println(err)
		session.Rollback()
		return false, session
	}
	return true, session
}
/**
 * 添加会员
 * @return    *User 会员信息
 * @return    error 错误信息
 */
func (self *User) Insert() (*User, error) {
	_, err := engine.InsertOne(self)
	return self, err
}

/**
 * 通过uid获取会员信息
 * @param uid int 会员id
 * @return    *User 会员信息
 * @return    error 错误信息
 */
func (User) GetUserByID(uid int) (*User, error) {
	var user User
	has, err := engineSlave.Where("uid=?", uid).Get(&user)
	if err != nil || !has {
		fmt.Println(err)
		return nil, err
	}
	return &user, nil
}

func (self *User) Update2() bool {
	_, err := engine.Id(self.Uid).Update(self)
	if err != nil {
		return false
	}
	return true
}

/**
 * 更新会员信息 ，通过map需要制定表
 * @param uid int 会员id
 * @return    bool
 */
func (User) Update(uid int, data map[string]interface{}) bool {
	res, err := engine.Table("lb_user").Where("uid=?", uid).Update(data)
	if res == 0 || err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

/**
 * 获取代理商的玩家数
 * @param uid int 会员id
 * @return    bool
 */
func (User) AgentUserNums(agent_id int) int {
	user := new(User)
	total, _ := engineSlave.Where("agent_id=? and account_state=?", agent_id, 1).Count(user)

	return int(total)
}
