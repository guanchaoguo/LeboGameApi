package config

//mongodb
type common struct {
	GAME_API_SUF         string //#游戏api后缀
	AGENT_IPS            string //代理商ip列表标识
	AGENT_ORDER_LIST     string //获取代理商用户注单信息后缀
	GET_ORDER_LIST_SPEED int64  //限定获取注单信息接口的时间间隔：秒
	GET_ORDER_LIST_MAX   int    //限定每次获取注单数据最大记录的条数
	MAX_TIME_SPAN        int64  //限定获取注单信息的最大时间跨度：天
	KEY_MAX_VALID_TIME   int    //SecurityKey最大有效时间:天
	SECURITY_KEY_ENCRYPT string //SecurityKey加密方式，默认为：sha1,可选项为sha256
	GAME_HOST            string //游戏客户端地址
	GAME_HOST_PC         string //游戏PC端地址
	TEST_USER_COUNT      int    //测试联调账号只能获取10条注单数据
	WHITELIST            string //代理商白名单redis key名称
	GAME_MT_ING          string //系统维护redis key
}


//mysql
type mysql struct {
	DB_HOST         string //数据库服务器
	DB_DATABASE     string //数据库名称
	DB_USERNAME     string //数据库登录名
	DB_PASSWORD     string //数据库密码
	DB_PORT         string //数据库端口
	CHARSET         string //字符集
	SetMaxIdleConns int    //默认打开数据库的连接数
	SetMaxOpenConns int    //最大打开数据库的连接数
}
//主库
type mysqlMaster struct {
	DB_MASTER mysql
}
//从库
type mysqlSlave struct {
	DB_SLAVE mysql
}
//mongodb
type mongodb struct {
	MONGO_DATABASE  string
	MONGO_URL       string
}

//redis
type redis struct {
	REDIS_HOST     string
	REDIS_PASSWORD string
	REDIS_PORT     string
}

type Default struct {
	REDIS_DEFAULT redis
}
type Monitor struct {
	REDIS_MONITOR redis
}
type Account struct {
	REDIS_ACCOUNT redis
}



