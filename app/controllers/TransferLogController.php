<?php
/**
 * Created by PhpStorm.
 * User: chensongjian
 * Date: 2017/3/21
 * Time: 13:07
 */

namespace App\controllers;

use bootstrap\Bootstrap;
use bootstrap\Input;

class TransferLogController
{
    function  __construct()
    {
        IpLimit::init();//ip白名单过滤
        SecurityKeyMiddleware::handle();//SecurityKey过滤
    }

    /**
     * @SWG\Post(
     *   path="/transferLog",
     *   tags={"api"},
     *   summary="会员存取款状态查询",
     *   description="
     * 当游戏接入商调用会员存款或取款接口，数据已发出但未接收到处理结果，
     * 则可使用此接口至游戏供应商查询处理状态。目前只支持单条记录查询。
     * 成功返回结果字段描述：
        {
            'code': 0,//状态码，0：成功，非0：错误
            'text': 'Success',//文本描述
            'result': ''//结果
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
     *     name="serial",
     *     type="string",
     *     description="流水号",
     *     required=true,
     *     default="LA222543362879271"
     *   ),
     *   @SWG\Parameter(
     *     in="formData",
     *     name="token",
     *     type="string",
     *     description="SHA1('securityKey|serial|agent')",
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
        $apiLog['apiName'] = '会员存取款状态查询';
        $apiLog['ip_info'] = get_real_ip();
        $apiLog['log_type'] = 'api';

        $post = Input::post();

        $agent = $post['agent'];
        $serial = $post['serial'];
        $token = $post['token'];

        // 联调状态统计 只要有联调请求则统计数据
        $debugging = false;
        if(ApiStatistics($apiLog)) $debugging = true;


        if( empty($agent)  || empty($serial) || empty($token)) {

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

       /* $key_name = $agent.config('common.AGENT_IPS');
        $white_list = json_decode(Bootstrap::Redis()->get($key_name), true);*/

        $is_auth = checkEequest($token,[getSecurityKey( $agent ), $serial, $agent]);

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

        $pkey = md5($agent.$serial.config('common.GAME_API_SUF'));
        $cashModel = Bootstrap::MongoDB()->selectCollection('live_game','cash_record');
        $res = $cashModel->findOne(['pkey'=> $pkey]);

        if( ! $res ) {

            $apiLog['code'] = 1008;
            $apiLog['text'] = config('errorcode.1008');
            $apiLog['result'] = '';
            $apiLog['end_time'] = time();
            addApiLog($apiLog);//日志记录

            return returnJson([
                'code'=> 1008,
                'text'=> config('errorcode.1008'),
                'result'=>'',
            ]);

        }

        $apiLog['code'] = 0;
        $apiLog['text'] = config('errorcode.9999');
        $apiLog['result'] = '';
        $apiLog['end_time'] = time();
        addApiLog($apiLog);//日志记录
        if($debugging) ApiSucceds($apiLog);//联调账号联调成功次数统计

        return returnJson([
            'code'=> 0,
            'text'=> config('errorcode.9999'),
            'result'=>'',
        ]);
    }
}