<?php
/**
 * Created by PhpStorm.
 * User: liangxz@szljfkj.com
 * Date: 2017/3/20
 * Time: 15:19
 */

namespace App\controllers;


use bootstrap\Bootstrap;
use bootstrap\Input;
use lib\db\DB;

class CashController
{
    public function __construct()
    {
        IpLimit::init();//白名单过滤
        SecurityKeyMiddleware::handle();//SecurityKey过滤

    }
    /**
     * @SWG\Post(
     *   path="/deposit",
     *   tags={"api"},
     *   summary="会员充值操作",
     *   description="
     * 会员将接入商钱包的钱，部分或者全部提取到游戏供应商的账户上进行游戏
     *
     * code： 0 成功；
     * text ：'接口错误信息描述'；
     * result [
     *          order_sn：充值订单号;
     *          amount：充值成功的金额;
     *      ]
     *     ",
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
     *     default="csj_play"
     *   ),
     *   @SWG\Parameter(
     *     in="formData",
     *     name="amount",
     *     type="number",
     *     description="存款金额",
     *     required=true,
     *     default="100.00"
     *   ), 
     *  @SWG\Parameter(
     *     in="formData",
     *     name="deposit_type",
     *     type="number",
     *     description="异常类型 当需要异常取消派彩充值使用",
     *     required=false,
     *     default="2"
     *   ),
     *   @SWG\Parameter(
     *     in="formData",
     *     name="token",
     *     type="string",
     *     description="SHA1('securityKey|username|amount|agent')",
     *     required=true,
     *     default="bc17e4f2734ee620bc4be2da0dfd868fbe8b1f72"
     *   ),
     *   @SWG\Response(response="200", description="Success")
     * )
     */
    public function deposit()
    {

        //var_export(sha1('32768aeb6c527875dc753c24fb1fcfb199045cdd|100|csj_play|agent_test'));die;
        $apiLog['start_time'] = time();
        $apiLog['user_name'] = Input::post('agent');
        $apiLog['postData'] = json_encode($_REQUEST);
        $apiLog['apiName'] = '会员充值';
        $apiLog['ip_info'] = get_real_ip();
        $apiLog['log_type'] = 'api';

        //先进行数据验证，防止数据被篡改
        $token = Input::post('token');
        $param = [getSecurityKey( Input::post('agent')), Input::post('username'), Input::post('amount'), Input::post('agent')];
        $amount = Input::post('amount');

        // 联调状态统计 只要有联调请求则统计数据
        $debugging = false;
        if(ApiStatistics($apiLog)) $debugging = true;

        //判断系统是否在维护当中
        if( $maintain = IpLimit::checkIsMaintain(Input::post('username'), Input::post('agent')) )
        {
            return $maintain;
        }

        $DB = new DB();
        $agent_ = $DB->row("SELECT id,agent_code ,parent_id FROM lb_agent_user WHERE user_name = :user_name", ['user_name' => Input::post('agent')]);

        $Prefix = substr(Input::post('agent'),0,2);
        $decry_username = Crypto::decrypt($agent_['agent_code'].Input::post('username'));

        
        $deposit_type = (int) Input::post('deposit_type') ?  Input::post('deposit_type') : 1;
        if($deposit_type == 2 ){
            array_push($param, $deposit_type);
            $apiLog['apiName'] = '会员取消派彩异常充值';
            $cashData['type'] = 10; //
        }

        //数据验证不通过，证明数据以及该被篡改过
        if(!checkEequest($token,$param))
        {
            return returnError();
        }

        //验证金钱格式类型
        if(!$this->makeValidate())//参数类型错误
        {
            $apiLog['code'] = 9002;
            $apiLog['text'] = config('errorcode.9002');
            $apiLog['result'] = '';
            $apiLog['end_time'] = time();
            addApiLog($apiLog);//日志记录

            return returnJson([
                'code'  => 9002,
                'text'  => config('errorcode.9002'),
                'result'    => ''
            ]);
        }

        if( $amount <= 0) {
            return returnJson([
                'code'  => 9002,
                'text'  => config('errorcode.9002'),
                'result'    => ''
            ]);
        }

        //判断会员是否存在
        //$user = Player::where('user_name',Input::post('username'))->first();
//        $user = Bootstrap::DB()->get('lb_user',"*",['user_name'=>Input::post('username'),'agent_name'=>Input::post('agent')]);
        $DB = new DB();
        $user = $DB->row("SELECT * FROM lb_user WHERE user_name = :user_name AND agent_name = :agent_name", ['user_name'=>$decry_username,'agent_name'=>Input::post('agent')]);
        if(!$user)
        {
            $apiLog['code'] = 1007;
            $apiLog['text'] = config('errorcode.1007');
            $apiLog['result'] = '';
            $apiLog['end_time'] = time();
            addApiLog($apiLog);//日志记录

            return returnJson([
                'code'  => 1007,
                'text'  => config('errorcode.1007'),
                'result'    => ''
            ]);
        }

        //进行会员存款操作
        $order_sn = createOrderSn();
        $cashModel = Bootstrap::MongoDB()->selectCollection('live_game','cash_record');
        $amount = sprintf("%.2f", $amount);
        $cashData['order_sn'] =$order_sn;
        $cashData['uid'] = (int)$user['uid'];
        $cashData['user_name'] = Crypto::encrypt($user['user_name']);
        $cashData['type'] =  $deposit_type == 2 ? $cashData['type'] : 1;
        $cashData['amount'] = (double) sprintf("%.2f",Input::post('amount'));
        $cashData['status'] = 3;//添加操作
        $cashData['user_money'] = (double)($user['money'] + $amount);
        $cashData['desc']    = '流水号：'.$order_sn;
        $cashData['admin_user'] = 'system-api';
        $cashData['admin_user_id'] = 0;
        $cashData['cash_no'] = $order_sn;
        $cashData['agent_id'] = (int)$agent_['id'];
        $cashData['hall_id'] = (int)$agent_['parent_id'];
        $cashData['add_time'] = new \MongoDB\BSON\UTCDateTime(time() * 1000);

//        $cashModel->pkey = md5($user->uid.$order_sn.$request->input('amount'));
        $cashData['pkey'] = md5(Input::post('agent').$order_sn.config('common.GAME_API_SUF'));
        //进行添加记录操作
        $res = $cashModel->insertOne($cashData);

        //充值成功需要修改用户的余额
        $userData['money'] = $user['money'] + $amount;
        if($res->isAcknowledged())
        {
            // 发起扣款
            $where = ['user_name'=>$decry_username,'agent_name'=>Input::post(agent),'money' => $amount];
            $update = $DB->query('UPDATE lb_user SET grand_total_money = money + :money, money = money + :money  '.
                         'WHERE user_name = :user_name AND agent_name = :agent_name', $where);
        }

        if(!$res->isAcknowledged() || !$update)
        {
            $apiLog['code'] = 3033;
            $apiLog['text'] = config('errorcode.3003');
            $apiLog['result'] = '';
            $apiLog['end_time'] = time();
            addApiLog($apiLog);//日志记录

            return returnJson([
                'code'  => 3003,
                'text'  => config('errorcode.3003'),
                'result'    => ''
            ]);
        }

        //用户增加充值、扣款后累计清除下注次数
        Bootstrap::Redis('monitor')->set("betcount:".$user['uid'], 0);

        $apiLog['code'] = 0;
        $apiLog['text'] = config('errorcode.9999');
        $apiLog['result'] = json_encode([ 'order_sn'  => $order_sn, 'amount'    => $amount]);
        $apiLog['end_time'] = time();
        addApiLog($apiLog);//日志记录
        if($debugging)  ApiSucceds($apiLog);//联调账号联调成功次数统计

        //统计充值记录
        $this->totalScoreRecord($user['agent_id'], $amount);
        //添加成功
        return returnJson([
            'code'  => 0,
            'text'  => config('errorcode.9999'),
            'result'    => [
                'order_sn'  => $order_sn,
                'amount'    => $amount
            ]
        ]);
    }

    /**
     * 玩家充值时 给代理统计
     * @param int $agent_id 代理商id
     * @param float $money 金额
     * @return int
     */
    private function totalScoreRecord( int $agent_id, $money) : int
    {
        $DB = new DB();
        $agent = $DB->row("SELECT * FROM lb_agent_user WHERE id = :agent_id", ['agent_id'=>$agent_id]);

        // 如果是不是停用账号 或者正式账号才统计数据
        if($agent['account_state'] != 1 || $agent['account_type'] != 1 ) return -1 ;

        if($agent['parent_id']){
            $hall = $DB->row("SELECT * FROM lb_agent_user WHERE id = :agent_id", ['agent_id'=>$agent['parent_id'] ]);
        }

        if( $agent ) {

            $where = [
                // 'day_year' => date('Y', time()),
                // 'day_month' => date('m', time()),
                // 'day_day' => date('d', time()),
                'add_date' => date('Y-m-d'),
                'agent_id' => $agent['id'],
            ];

            $re = $DB->row("SELECT * FROM statis_cash_agent WHERE  add_date = :add_date AND agent_id = :agent_id", $where);

            if( ! $re ) {
                $where = [
                    'day_year' => date('Y', time()),
                    'day_month' => date('m', time()),
                    'day_day' => date('d', time()),
                    'agent_id' => $agent['id'],
                ];

                $where['add_date'] = date('Y-m-d', time());
                $where['hall_id'] = $agent['parent_id'];
                $where['agent_name'] = $agent['user_name'];
                $where['hall_name'] = $hall['user_name'];
                $where['total_score_record'] = $money;

                $res = $DB->insert('statis_cash_agent',$where);
                if( $res ) {
                    return 1;
                } else {
                    return -1;
                }
            } else {
                $where['money'] = $money;
                $res = $DB->query("UPDATE statis_cash_agent SET total_score_record = total_score_record + :money WHERE add_date = :add_date  AND agent_id = :agent_id", $where);

                if( $res !== false) {
                    return 1;
                } else {
                    return -1;
                }
            }
        } else {
            return -1;
        }

    }

    /**
     * 数据验证操作
     * @return bool
     */
    private function makeValidate()
    {
        $amount = Input::post('amount');
        $agent = Input::post('agent');
        $username = Input::post('username');

        if(empty($amount) || empty($amount) || empty($username) || !is_numeric($amount))
            return false;

        // 限制单次充值金额不超过1亿
        if($amount > (100000000 -1))
            return false;

        return true;
    }




}