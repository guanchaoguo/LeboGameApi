<?php

namespace App\controllers;

use bootstrap\Bootstrap;
use bootstrap\Input;
use lib\db\DB;

class PlayerController
{

    function  __construct()
    {
        IpLimit::init();//ip白名单过滤
        SecurityKeyMiddleware::handle();//SecurityKey过滤
    }

    /**
     * @SWG\Post(
     *   path="/user",
     *   tags={"api"},
     *   summary="获取供应商会员信息",
     *   description="
     * 获取会员在供应商的相关信息
     * 成功返回结果字段描述：
        {
            'code': 0,//状态码，0：成功，非0：错误
            'text': 'Success',//文本描述
            'result': {//结果
                'online_status': 'Online',//在线状态，Online：在线，Offline：离线
                'member_status': 'Normal',//会员状态，Normal：正常，Abnormal：异常
                'balance': '1990.00'//余额
            }
        }",
     *   operationId="index",
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
     *     name="token",
     *     type="string",
     *     description="SHA1('securityKey|username|agent')",
     *     required=true,
     *     default="74b923d6f6ab1d1dc450eb72158759e4f1f964da"
     *   ),
     *   @SWG\Response(response="200", description="成功")
     * )
     */
    public function index()
    {
        $apiLog['start_time'] = time();
        $apiLog['user_name'] = Input::post('agent');
        $apiLog['postData'] = json_encode($_REQUEST);
        $apiLog['apiName'] = '获取供应商会员信息';
        $apiLog['ip_info'] = get_real_ip();
        $apiLog['log_type'] = 'api';

        $post = Input::post();

        $agent = $post['agent'];
        $username = $post['username'];
        $token = $post['token'];
        $Prefix = substr($agent,0,2);

        // 联调状态统计 只要有联调请求则统计数据
        $debugging = false;
        if(ApiStatistics($apiLog)) $debugging = true;

        if( empty($agent)  || empty($username) || empty($token)) {
            $apiLog['code'] = 9002;
            $apiLog['text'] = config('errorcode.9002');
            $apiLog['result'] = '';
            $apiLog['end_time'] = time();
            addApiLog($apiLog);//日志记录

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

        $is_auth = checkEequest($token,[getSecurityKey( $agent ), $username, $agent]);

        if(!$is_auth) {

            $apiLog['code'] = 9002;
            $apiLog['text'] = config('errorcode.9002');
            $apiLog['result'] = '';
            $apiLog['end_time'] = time();
            addApiLog($apiLog);//日志记录

            return returnJson([
                'code'=> 9002,
                'text'=> config('errorcode.9002'),
                'result'=>'',
            ]);

        }
        $DB = new DB();



        $agent_ = $DB->row("SELECT id,agent_code FROM lb_agent_user WHERE user_name = :user_name", ['user_name' => $agent]);

        $where = [
            'user_name' => Crypto::decrypt($agent_['agent_code'].$username),
            'agent_name' => $agent,
        ];
        $lb_user = $DB->row("SELECT uid,money,on_line,account_state FROM lb_user WHERE user_name = :user_name AND agent_name = :agent_name", $where);

//        $lb_user = Bootstrap::DB()->get('lb_user', ['uid','money','on_line','account_state'], $where);

        if(!$lb_user) {

            $apiLog['code'] = 1007;
            $apiLog['text'] = config('errorcode.1007');
            $apiLog['result'] = '';
            $apiLog['end_time'] = time();
            addApiLog($apiLog);//日志记录

            return returnJson([
                'code'=> 1007,//用户不存在
                'text'=> config('errorcode.1007'),
                'result'=>'',
            ]);

        }


        $apiLog['code'] = 0;
        $apiLog['text'] = config('errorcode.9999');
        $apiLog['result'] = json_encode(['online_status' => $lb_user['on_line'] == 'Y' ? 'Online' : 'Offline','member_status' => $lb_user['account_state'] == 1 ? 'Normal' : 'Abnormal','member_status' => $lb_user['account_state'] == 1 ? 'Normal' : 'Abnormal']);
        $apiLog['end_time'] = time();
        addApiLog($apiLog);//日志记录
        if($debugging) ApiSucceds($apiLog);//联调账号联调成功次数统计

        return returnJson([
            'code'=> 0,
            'text'=> config('errorcode.9999'),
            'result'=>[
                'online_status' => $lb_user['on_line'] == 'Y' ? 'Online' : 'Offline',
                'member_status' => $lb_user['account_state'] == 1 ? 'Normal' : 'Abnormal',
                'balance'       => $lb_user['money'],
            ],
        ]);

    }

    /**
     * @SWG\Post(
     *   path="/withDrawal",
     *   tags={"api"},
     *   summary="会员取款",
     *   description="
     *  会员将游戏供应商的账户上金额转存到接入商，
     *  取款时需调用“检查用户状态”接口检查用户金额避免错误。
     *  成功返回结果字段描述：
        {
            'code': 0,//状态码，0：成功，非0：错误
            'text': 'Success',//文本描述
            'result': {//结果
                'amount': '10',//扣取金额
                'order_sn': 'LA321867200326457'//流水号
            }
        }",
     *   operationId="withDrawal",
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
     *     name="amount",
     *     type="number",
     *     description="金额",
     *     required=true,
     *     default="1.00"
     *   ),
     *   @SWG\Parameter(
     *     in="formData",
     *     name="token",
     *     type="string",
     *     description="SHA1('securityKey|username|amount|agent')",
     *     required=true,
     *     default="74b923d6f6ab1d1dc450eb72158759e4f1f964da"
     *   ),
     *   @SWG\Response(response="200", description="成功")
     * )
     */
    public function withDrawal()
    {
        $apiLog['start_time'] = time();
        $apiLog['user_name'] = Input::post('agent');
        $apiLog['postData'] = json_encode($_REQUEST);
        $apiLog['apiName'] = '会员取款';
        $apiLog['ip_info'] = get_real_ip();
        $apiLog['log_type'] = 'api';

        $post = Input::post();

        $agent = $post['agent'];
        $username = $post['username'];
        $amount = $post['amount'];
        $token = $post['token'];
        $Prefix = substr($agent,0,2);

        // 联调状态统计 只要有联调请求则统计数据
        $debugging = false;
        if(ApiStatistics($apiLog)) $debugging = true;

        if( empty($agent)  || empty($username) || empty($amount) || !is_numeric($amount) || $amount<=0 || empty($token) ) {

            $apiLog['code'] = 9002;
            $apiLog['text'] = config('errorcode.9002');
            $apiLog['result'] = '';
            $apiLog['end_time'] = time();
            addApiLog($apiLog);//日志记录

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

        $is_auth = checkEequest($token,[getSecurityKey( $agent ), $username, $amount, $agent]);

        if(!$is_auth) {

            $apiLog['code'] = 9002;
            $apiLog['text'] = config('errorcode.9002');
            $apiLog['result'] = '';
            $apiLog['end_time'] = time();
            addApiLog($apiLog);//日志记录

            return returnJson([
                'code'=> 9002,
                'text'=> config('errorcode.9002'),
                'result'=>'',
            ]);

        }
        $DB = new DB();
        $agent_ = $DB->row("SELECT id,agent_code,parent_id FROM lb_agent_user WHERE user_name = :user_name", ['user_name' => $agent]);

        $lb_user = $DB->row("SELECT * FROM lb_user WHERE user_name = :user_name AND agent_name = :agent_name", ['user_name' => Crypto::decrypt($agent_['agent_code'].$username), 'agent_name' => $agent]);
//        $lb_user = Bootstrap::DB()->get('lb_user', '*', ['user_name' => $username, 'agent_name' => $agent]);

        if(! $lb_user ) {

            $apiLog['code'] = 1007;
            $apiLog['text'] = config('errorcode.1007');
            $apiLog['result'] = '';
            $apiLog['end_time'] = time();
            addApiLog($apiLog);//日志记录

            return returnJson([
                'code'=> 1007,//用户不存在
                'text'=> config('errorcode.1007'),
                'result'=>'',
            ]);
        }

        if( ($lb_user['money'] - $amount) < 0) {

            $apiLog['code'] = 3001;
            $apiLog['text'] = config('errorcode.3001');
            $apiLog['result'] = '';
            $apiLog['end_time'] = time();
            addApiLog($apiLog);//日志记录

            return returnJson([
                'code'=> 3001,//余额不足
                'text'=> config('errorcode.3001'),
                'result'=>'',
            ]);

        }

        // 发起扣款
        $re = $DB->query("UPDATE lb_user SET grand_total_money = money - :money, money = money - :money WHERE uid = :uid AND money > 0", ['uid' => $lb_user['uid'],'money' => $amount]);

        if( ! $re  ) {

            $apiLog['code'] = 3002;
            $apiLog['text'] = config('errorcode.3002');
            $apiLog['result'] = '';
            $apiLog['end_time'] = time();
            addApiLog($apiLog);//日志记录

            return returnJson([
                'code'=> 3002,//取款失败
                'text'=> config('errorcode.3002'),
                'result'=>'',
            ]);
        }

        $lb_user = $DB->row("SELECT * FROM lb_user WHERE user_name = :user_name AND agent_name = :agent_name", ['user_name' => Crypto::decrypt($agent_['agent_code'].$username), 'agent_name' => $agent]);

        $ordernum = createOrderSn();
        $cashData['order_sn'] = $ordernum;
        $cashData['uid'] = (int)$lb_user['uid'];
        $cashData['user_name'] = Crypto::encrypt($lb_user['user_name']);
        $cashData['type'] = 1;//转账
        $cashData['amount'] = (double) $amount;
        $cashData['status'] = 4;//扣取
        $cashData['user_money'] = (double)($lb_user['money']);
        $cashData['desc'] = '流水号：'.$ordernum;
        $cashData['admin_user'] = 'system-api';
        $cashData['admin_user_id'] = 0;
        $cashData['cash_no'] = $ordernum;
        $cashData['agent_id'] = (int)$agent_['id'];
        $cashData['hall_id'] = (int)$agent_['parent_id'];
        $cashData['add_time'] = new \MongoDB\BSON\UTCDateTime(time() * 1000);
        $cashData['pkey'] =  md5($agent.$ordernum.config('common.GAME_API_SUF'));

        $cashModel = Bootstrap::MongoDB()->selectCollection('live_game','cash_record');
        $res = $cashModel->insertOne($cashData);

        if($res->isAcknowledged()){

            //用户增加充值、扣款后累计清除下注次数
            Bootstrap::Redis('monitor')->set("betcount:".$lb_user['uid'], 0);

            $apiLog['code'] = 0;
            $apiLog['text'] = config('errorcode.9999');
            $apiLog['result'] = '';
            $apiLog['end_time'] = time();
            addApiLog($apiLog);//日志记录
            if($debugging)  ApiSucceds($apiLog);//联调账号联调成功次数统计

            return returnJson([
                'code' => 0,
                'text' => config('errorcode.9999'),
                'result' => [
                    'amount' => $amount,
                    'order_sn' => $ordernum,
                ],
            ]);
        }

        $apiLog['code'] = 3002;
        $apiLog['text'] = config('errorcode.3002');
        $apiLog['result'] = '';
        $apiLog['end_time'] = time();
        addApiLog($apiLog);//日志记录

        return returnJson([
            'code' => 3002,//取款失败
            'text' => config('errorcode.3002'),
            'result' => '',
        ]);


    }
}
