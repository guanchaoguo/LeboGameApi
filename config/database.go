package config

import (
	"encoding/json"
	io "io/ioutil"
	"os/exec"
	"os"
	"path/filepath"
)

var conf []byte

var def Default
var monitor Monitor
var account Account

func init()  {
	//获取配置文件信息
	getConf()
	//获取redis json 配置
	initRedis()
}

func getConf()  {
	//获取当前执行文件的路径
	file, _ := exec.LookPath(os.Args[0])
	AppPath, _ := filepath.Abs(file)
	path, _ := filepath.Split(AppPath)
	_, err := os.Stat(path + "/conf.json")
	confPath := "./conf.json"
	if	err == nil {
		confPath = path + "/conf.json"
	}
	//初始化配置
	data, err := io.ReadFile(confPath)
	if err != nil{
		return
	}
	conf = []byte(data)
}

//获取redis 配置（由于reids没有连接池，所以启动时就获取json配置信息）
func initRedis()  {
	//默认 redis
	def = Default{}
	json.Unmarshal(conf,&def)
	//monitor redis
	monitor = Monitor{}
	json.Unmarshal(conf,&monitor)
	//account redis
	account = Account{}
	json.Unmarshal(conf,&account)
}
/*
	数据库配置文件
*/
func Common() common {
	com := common{}
	json.Unmarshal(conf,&com)
	return com
}

//mysql配置
func GetMysqlConf(connect string) mysql {

	switch connect {
	case "master":
		db_master := mysqlMaster{}
		json.Unmarshal(conf,&db_master)
		return db_master.DB_MASTER
	case "slave":
		db_slave := mysqlSlave{}
		json.Unmarshal(conf,&db_slave)
		return db_slave.DB_SLAVE
	default:
		db_master := mysqlMaster{}
		json.Unmarshal(conf,&db_master)
		return db_master.DB_MASTER
	}
	//db_mysql := mysql{}
	//json.Unmarshal(conf,&db_mysql)
	//return db_mysql
}

//mongodb配置
func GetMongodb() (string, string) {
	db_mongodb := mongodb{}
	json.Unmarshal(conf,&db_mongodb)
	return db_mongodb.MONGO_URL, db_mongodb.MONGO_DATABASE
	//url := "mongodb://hhq163:bx123456@192.168.31.231:27017/" + dbname + "?connect=direct&maxPoolSize=100"
	//url := "mongodb://hhq163:bx123456@192.168.31.233:30000,192.168.31.234:30000,192.168.31.235:30000/"+ dbname
	//mongo, err = mgo.Dial("mongodb://hhq163:bx123456@192.168.31.233:30000,192.168.31.234:30000,192.168.31.235:30000/live_game?connect=&maxPoolSize=")
	//return url, dbname
}

//redis配置
func GetRedis(connect string) redis {

	switch connect {
	case "default": //代理商开通游戏种类关系表,每个代理的限额,代理商名称和厅主白名单关系,域名管理，系统维护
		return def.REDIS_DEFAULT
	case "monitor": //风险监控,用户增加充值、扣款后累计下注次数,用户/ip刷水统计
		return monitor.REDIS_MONITOR

	case "account": //已登录过用户数据
		return account.REDIS_ACCOUNT

	default:
		return def.REDIS_DEFAULT

	}

}
