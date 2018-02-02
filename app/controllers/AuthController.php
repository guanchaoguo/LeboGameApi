<?php

namespace App\controllers;

use bootstrap\Bootstrap;
use bootstrap\Input;
use lib\db\DB;
class AuthController
{

    function  __construct()
    {

        IpLimit::init();//ip白名单过滤

        SecurityKeyMiddleware::handle();//SecurityKey过滤/

    }

    /**
     * consumes={"multipart/form-data"},
     * @SWG\Post(
     *   path="/authorization",
     *   tags={"api"},
     *   summary="会员登录&注册",
     *   description="
     * 会员登录到供应商进行游戏。首先供应商会检测用户是否存在，
     * 如果不存在，供应商自动创建账号，如果存在则返回游戏地址，直接进入游戏大厅，其它错误提示如下。
     * 成功返回结果字段描述：
    {
    'code': 0,//状态码，0：成功，非0：错误
    'text': 'Success',//文本描述
    'result': 'http://www.alibaba.com?uid=321d1aee14044e1a07193'//游戏url地址
    }",
     *
     *   operationId="login",
     *   @SWG\Parameter(
     *     in="formData",
     *     name="agent",
     *     type="string",
     *     description="代理商用户名",
     *     required=true,
     *     default="agent_test"
     *   ),
     *   @SWG\Parameter(
     *     in="formData",
     *     name="username",
     *     type="string",
     *     description="用户登录名称",
     *     required=true,
     *     default="chentest"
     *   ),
     *   @SWG\Parameter(
     *     in="formData",
     *     name="login_type",
     *     type="number",
     *     description="登录类型：1：pc，2:h5，默认为pc",
     *     required=true,
     *     default="1"
     *   ),
     *   @SWG\Parameter(
     *     in="formData",
     *     name="account_type",
     *     type="number",
     *     description="玩家类型：2 试玩玩家",
     *     required=false,
     *     default="1"
     *   ),
     *   @SWG\Parameter(
     *     in="formData",
     *     name="token",
     *     type="string",
     *     description="SHA1('securityKey|username|agent|login_type')/SHA1('securityKey|account_type|agent|login_type') ",
     *     required=true,
     *     default="74b923d6f6ab1d1dc450eb72158759e4f1f964da"
     *   ),
     *   @SWG\Response(response="200",
     *     description="Success"
     * )
     * )
     */
    public function login()
    {
        //日志捕获
        $apiLog['start_time'] = time();
        $apiLog['agent'] = Input::post('agent'); //代理商名称
        $apiLog['postData'] = json_encode($_REQUEST);
        $apiLog['apiName'] = '玩家登录游戏';
        $apiLog['ip_info'] = get_real_ip();
        $apiLog['log_type'] = 'login';// 登录日志
        $apiLog['user_name'] = Input::post('username'); 
        
        $post = Input::post();

        $agent = $post['agent'];
        $username = $post['username'];
        $token = $post['token'];

        //登录类型
        $login_type = (int)$post['login_type'] ? $post['login_type'] : 1;

        // 联调状态统计 只要有联调请求则统计数据
        $debugging = false;
        if(ApiStatistics($apiLog)) $debugging = true;

        if( ! in_array($login_type, [1, 2]) ) {
            return returnJson([
                'code' => 9002,
                'text' => config('errorcode.9002'),
                'result' => '' ,
            ]);
        }
//        $Prefix = substr($agent,0,2);

        if(!isset($post['account_type'])){
            if( ! preg_match('/^[a-zA-Z0-9_]{3,20}$/', $username) ) {

                $apiLog['code'] = 9002;
                $apiLog['text'] = config('errorcode.9002');
                $apiLog['result'] = '';
                $apiLog['end_time'] = new \MongoDB\BSON\UTCDateTime(time() * 1000);
                addApiLog($apiLog);//日志记录

                return returnJson([
                    'code' => 9002,
                    'text' => config('errorcode.9002'),
                    'result' => '' ,
                ]);
            }
        }

        //账户类型
        $account_type = (int)$post['account_type'] ? $post['account_type'] : 1;
        if( ! in_array($account_type, [1, 2]) ) {
            return returnJson([
                'code' => 9002,
                'text' => config('errorcode.9002'),
                'result' => '' ,
            ]);
        }

        //判断系统是否在维护当中
        if( $maintain = IpLimit::checkIsMaintain(Input::post('username'), Input::post('agent')) )
        {
            return $maintain;
        }


        /*$key_name = $agent.config('common.AGENT_IPS');
        $white_list = json_decode(Bootstrap::Redis()->get($key_name), true);*/
        // 兼容2.01版本
        $securityKey = getSecurityKey( $agent );
        $checktoken = [$securityKey, $username, $agent,$login_type];
        if($account_type == 2 ){
            $checktoken = [$securityKey, $account_type, $agent,$login_type];
        }
        $is_auth = checkEequest($token,$checktoken);

        if(!$is_auth) {

            $apiLog['code'] = 9002;
            $apiLog['text'] = config('errorcode.9002');
            $apiLog['result'] = '';
            $apiLog['end_time'] = time();
                     
            addApiLog($apiLog);//登录日志记录

            return returnJson([
                'code'=> 9002,
                'text'=> config('errorcode.9002'),
                'result'=>'',
            ]);

        }

        $DB = new DB();

        $agent_ = $DB->row("SELECT id, parent_id, user_name, agent_code, account_type FROM lb_agent_user WHERE user_name = :user_name", ['user_name' => $agent]);

        $decry_username = Crypto::decrypt($agent_['agent_code'].$username);

        // 保存用户名
        $apiLog['user_name'] = $agent_['agent_code'].$username;

        $decry_username_md = Crypto::decrypt($username);
        $lb_user = $DB->row("SELECT * FROM lb_user WHERE user_name = :user_name", ['user_name' => $decry_username]);
        // $lb_user = Bootstrap::DB()->get('lb_user', '*', ['user_name' => $username]);
        // var_export($lb_user);die;

        // 创建玩家资料
        if( $account_type == 2   ||  (! $lb_user) ) {
            // 检查联调玩家数量是否超过限制
            if($agent_['account_type'] == 3 ){
                if(! $playerLimit =  self::checkVaildAgent($DB ,$agent_['id'])) {
                    return $playerLimit;
                }
            }

            //测试代理用户（免费试玩）不能通过api注册F
            if( $agent_['account_type'] == 2 ) {
                return returnJson([
                    'code' => 1005,//创建会员错误
                    'text' => config('errorcode.1003'),
                    'result' => '',
                ]);
            }

            $hall_ = $DB->row("SELECT id,user_name FROM lb_agent_user WHERE id = :id", ['id' => $agent_['parent_id']]);

            // 如果为试玩玩家则指定将玩家绑定到uid =2 的玩家代理上面
            $alias = 'api会员'; // 会员类型
            if($account_type == 2){
                // 查询玩家试玩代理名称
                $lb_user = $DB->row("SELECT * FROM lb_user WHERE uid = :uid", ['uid' => 2]);
                $userMoney = $lb_user['money'];

                // 查询试玩代理信息
                $agent_ = $DB->row(
                    "SELECT id, parent_id, user_name, agent_code, account_type FROM lb_agent_user WHERE user_name = :user_name",
                    ['user_name' => $lb_user['agent_name']]
                );

                // 查询试玩厅主
                $hall_ = $DB->row("SELECT id,user_name FROM lb_agent_user WHERE id = :id", ['id' => $agent_['parent_id']]);

                // 会员类型
                $alias = 'test member';

                // 随机唯一玩家用户名 和订单号同名
                $useRand = uniqid(mt_rand(0,99999999));
                $decry_username = Crypto::decrypt($agent_['agent_code'].$useRand);

                // 试玩玩家密码
                $decry_username_md = Crypto::decrypt($useRand);
            }

            $salt = randomkeys(20);
            $pwd = Crypto::decrypt(make_password());
            $attributes = [
                'user_name' => $decry_username,//玩家在第三方平台账号
//                'username_md' => $Prefix.$username,//玩家在平台的账号
                'username_md' => $decry_username_md,//玩家在平台的账号
                'password' => $pwd,//玩家在第三方平台密码
                'password_md' => $pwd,//玩家在平台的密码
                'alias' => $alias,
                'create_time' => date('Y-m-d H:i:s',time()),
                'add_date' => date('Y-m-d H:i:s',time()),
                'add_ip' => get_real_ip(),
                'ip_info' => get_real_ip(),
                'hall_id' => $hall_['id'],
                'hall_name' => $hall_['user_name'],
                'agent_id' => $agent_['id'],
                'agent_name' => $agent_['user_name'],
                'salt' => $salt,
                'user_rank'=> $agent_['account_type'] ==1 ? 0 : 1,
            ];
            $result = $DB->insert('lb_user',$attributes);

            if(!$user_id = $DB->lastInsertId()){

                $apiLog['code'] = 1005;
                $apiLog['text'] = config('errorcode.1005');
                $apiLog['result'] = '';
                $apiLog['end_time'] = time();

                addApiLog($apiLog);//登录日志记录

                return returnJson([
                    'code' => 1005,//创建会员错误
                    'text' => config('errorcode.1005'),
                    'result' => '',
                ]);

            }
            //如果为测试类型用户，则不进行账户数量的累加
            if($agent_['account_type'] ==1) {
                //更新代理商玩家数
                $DB->query("UPDATE lb_agent_user SET sub_user = sub_user+1 WHERE id = :id", ['id' => $agent_['id']]);
                //更新厅主代理商玩家数
                $DB->query("UPDATE lb_agent_user SET sub_user = sub_user+1 WHERE id = :id", ['id' => $hall_['id']]);
            }

            $lb_user = $DB->row("SELECT * FROM lb_user WHERE uid = :uid", ['uid' => $user_id]);
        }


        $account_state = $lb_user['account_state'];
        if($account_state != 1 && $account_state != 4) {

            $apiLog['code'] = 1001;
            $apiLog['text'] = config('errorcode.1001');
            $apiLog['result'] = '';
            $apiLog['end_time'] = time();
            addApiLog($apiLog);//登录日志记录

            return returnJson([
                'code'=> 1001, //账号停用
                'text'=> config('errorcode.1001'),
                'result'=>'',
            ]);

        }

        if(time() - strtotime($lb_user['last_time']) < 10) {
            $apiLog['code'] = 1004;
            $apiLog['text'] = config('errorcode.1004');
            $apiLog['result'] = '';
            $apiLog['end_time'] = time();
            addApiLog($apiLog);//登录日志记录

            return returnJson([
                'code'=> 1004,//登录频繁
                'text'=> config('errorcode.1004'),
                'result'=>'',
            ]);

        }

        //存redis，以供游戏端获取
//        $session_id = md5( $lb_user['user_name'].$lb_user['password'] );
        $session_id = md5( $lb_user['user_name'] );
        $uid = substr( $session_id, 0, 21 );
        $user_name = $lb_user['user_name'];
        $lb_user['user_name'] = Crypto::encrypt($user_name);
        $lb_user['username_md'] = Crypto::encrypt($lb_user['username_md']);
        $lb_user['username2'] = $user_name;
        $lb_user['time'] = time();
        $lb_user['agent_code'] = $agent_['agent_code'];

        Bootstrap::Redis('account')->set($uid, json_encode($lb_user));

        // 为试玩玩家充值
        if($account_type == 2){
            $this->onTrial( $lb_user, $apiLog ,$userMoney);
        }


        //更新数据
        $update_array = array(
            "on_line" => "Y",
//            "last_time" => date('Y-m-d H:i:s',time()),
            "ip_info" => get_real_ip(),
            "token_id" => $uid,
            "uid" => $lb_user['uid'],
        );

        if( $account_state == 4 ) {
            $update_array['account_state'] = 1;
        } else {
            $update_array['account_state'] = $account_state;
        }
        $DB->query("UPDATE lb_user SET on_line = :on_line,ip_info = :ip_info, token_id = :token_id,account_state = :account_state WHERE uid = :uid", $update_array);
//        Bootstrap::DB()->update('lb_user', $update_array, ['uid' => $lb_user['uid']]);

        $apiLog['code'] = 0;
        $apiLog['text'] = config('errorcode.9999');
        $apiLog['result'] = json_encode(['url'=>config('common.GAME_HOST_PC').'?uid='.$uid]);
        $apiLog['end_time'] =  time();
        
        addApiLog($apiLog);//日志记录
        if($debugging)  ApiSucceds($apiLog);//联调账号联调成功次数统计

        // 获取缓存中域名地址
        $gamehost = Bootstrap::Redis()->hget('GAMEHOST:URL',$login_type);

        switch ($login_type) {
            case 1:
                $gameDomain = $gamehost ? $gamehost.'/game.php': config('common.GAME_HOST_PC');
                $url = $gameDomain.'?uid='.$uid;
                break;
            case 2:
                $gameDomain = $gamehost ? $gamehost: config('common.GAME_HOST');
                $url = $gameDomain.'?uid='.$uid;
                break;
            default:
                $url = config('common.GAME_HOST_PC').'?uid='.$uid;

        }
        return returnJson([
            'code' => 0,
            'text' => config('errorcode.9999'),
            'result' => $url
        ]);
    }

    /**
      * 为试玩玩家充值
      */
    private function onTrial($user, $apiLog, $userMoney)
    {

        //修改用户的余额
        $update = (new DB())->query("UPDATE lb_user SET money = :money WHERE uid = :uid", ['money' => $userMoney ,'uid'=> $user['uid']]);
        if(!$update)
        {
            $apiLog['code'] = 3033;
            $apiLog['text'] = config('errorcode.3003');
            $apiLog['result'] = '';
            $apiLog['end_time'] = new \MongoDB\BSON\UTCDateTime(time() * 1000);
            addApiLog($apiLog);//日志记录

            return returnJson([
                'code'  => 3003,
                'text'  => config('errorcode.3003'),
                'result'    => ''
            ]);
        }

    }


    /**
     * 检查联调代理是否玩家超过给定限制
     */
    private  function checkVaildAgent($DB,$agent_id )
    {

        $playerNum = (int)config('common.TEST_USER_COUNT') ;// 联调账号玩家数量不超过10个
        $playerNum = $playerNum ? $playerNum : 10 ;

        // 统计代理的数量玩家数量
        $userNum  = $DB->row(" SELECT  COUNT(1) as num  FROM lb_user WHERE agent_id =:agent_id  AND account_state =:account_state ", ['agent_id' => $agent_id, 'account_state'=> 1 ]);
        if($userNum['num'] > $playerNum){
            return returnJson([
                'code'  => 4001,//联调玩家数量超过限制
                'text'  => config('errorcode.4001'),
                'result'   => ''
            ]);
        }

        return true;
    }


}
