<?php
/**
 * Created by PhpStorm.
 * User: liangxz@szljfkj.com
 * Date: 2017/3/20
 * Time: 10:01
 */
namespace App\controllers;

use bootstrap\Bootstrap;
use bootstrap\Input;
use lib\db\DB;

class KeyController
{

    public function __construct()
    {
        IpLimit::init();//IP白名单过滤
    }

    /**
     * @SWG\Post(
     *   path="/getKey",
     *   tags={"api"},
     *   summary="获取SecurityKey",
     *   description="
     * 接入商在进行其他业务操作时，必须要带上SecurityKey；SecurityKey是作为一个身份加密标识，而且是会过期，所以要通过接口形式获取；
     * 当调用该接口时，会生成一个新的SecurityKey返回，同时返回该SecurityKey的最后有效时间；接入商必须要存储SecurityKey和对应的最后过期时间，
     * 每次使用SecurityKey时都需要接入商先判断SecurityKey的过期时间，如果已经过期则需要重新调用该接口获取新的SecurityKey；
     * 注意：当调用该接口时，如果当前的SecurityKey是否已经过期，则会生成一个新的SecurityKey，覆盖原来的SecurityKey
     *
     * 接口信息返回说明：
     * code： 0 成功；
     * text ：'接口错误信息描述'；
     * result [
     *        security_key：生成的security_key
     *        expiration：security_key的最后有效时间
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
     *     name="token",
     *     type="string",
     *     description="SHA1('agent')",
     *     required=true,
     *     default="81b06f3e208548a4f8a949e0439afeb29139e1cb"
     *   ),
     *   @SWG\Response(response="200", description="Success")
     * )
     */
    public function createKey()
    {
        //var_export(hash('sha1','agent_test'));die;

        $apiLog['start_time'] = time();
        $apiLog['user_name'] = Input::post('agent');
        $apiLog['postData'] = json_encode($_REQUEST);
        $apiLog['apiName'] = '获取SecurityKey';
        $apiLog['ip_info'] = $_SERVER['SERVER_ADDR'];
        $apiLog['log_type'] = 'api';

        $param[] = Input::post('agent');


        // 联调状态统计 只要有联调请求则统计数据
        $debugging = false;
        if(ApiStatistics($apiLog)) $debugging = true;

        if (!checkEequest(Input::post('token'), $param))
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
        $agent = Input::post('agent');

        // 查询所属厅主信息
        if (!$agentInfo = getHallByAgent($agent)) {
            return returnJson([
                'code' => 9004,//代理所属的厅主不存在
                'text' => config('errorcode.9004'),
                'result' => '',
            ]);
        }

        // 获取缓存数据中的数据是否命中名单信息
        $agentKeyName = 'agentWhitelist'; $DB = new DB();
        if (!$ips = Bootstrap::Redis()->hGet($agentKeyName,$agent)) {
            // 未命中缓存获取数据库白名单信息
            $agentObj = $DB->row("SELECT * FROM white_list WHERE agent_id = :agent_id AND state = :state", ["agent_id" => $agentInfo['parent_id'], 'state' => 1]);
            $ips = json_encode($agentObj);
        }

        $agentObj =  json_decode($ips, true);
        if($agentObj && strtotime($agentObj['seckey_exp_date']) > time()) {
            $securityKey = $agentObj['agent_seckey'];
            $seckey_exp_date = $agentObj['seckey_exp_date'];
        } else {
            //SecurityKey生成规则：根据代理商用户名和mt_rand(10,100000)随机数拼接， 然后字符串打散处理，最后进行 sha1()加密返回 SecurityKey的有效时间为一个月

            // 创建key和有效时间
            $str = str_shuffle($agent.mt_rand(10,100000));
            $securityKey = $this->createSecurityKey(config('common.SECURITY_KEY_ENCRYPT'),$str);
            $validTime = (int)config('common.KEY_MAX_VALID_TIME');
            $seckey_exp_date = date('Y-m-d H:i:s',strtotime("+$validTime day"));
            
            //生成后进行修改数据数据库操作
            $setData = ['agent_seckey'=>$securityKey,'seckey_exp_date'=>$seckey_exp_date,"agent_id"=>$agentInfo['parent_id']];
            $res = $DB->query("UPDATE white_list SET agent_seckey = :agent_seckey,seckey_exp_date = :seckey_exp_date WHERE agent_id = :agent_id",$setData );
            if(!$res) {

                $apiLog['code'] = 9006;
                $apiLog['text'] = config('errorcode.9006');
                $apiLog['result'] = '';
                $apiLog['end_time'] = time();
                addApiLog($apiLog);//日志记录

                return returnJson([
                    'code' => 9006,
                    'text' => config('errorcode.9006'),
                    'result' => ''
                ]);
            }

            //修改成功进行redis存储缓存
            $agentObj = $DB->row("SELECT * FROM white_list WHERE agent_id = :agent_id AND state = :state", ["agent_id"=>$agentInfo['parent_id'], 'state' =>1]);
            // 类型转换
            $agentObj['id'] = (int)$agentObj['id'];
            $agentObj['agent_id'] = (int)$agentObj['agent_id'];
            $agentObj['state'] = (int)$agentObj['state'];

            // 额外字段
            $agentObj['agent_code'] = $agentInfo['agent_code'];
            $agentObj['account_type'] = (int)$agentInfo['account_type'];
            $agentObj['agent_id2'] = (int)$agentInfo['id'];
            $agentKeyName = 'agentWhitelist';
            Bootstrap::Redis()->hSet($agentKeyName,Input::post('agent'),json_encode($agentObj));

            // 重新赋值
            $securityKey = $agentObj['agent_seckey'];
            $seckey_exp_date = $agentObj['seckey_exp_date'];
        }


        $apiLog['code'] = 0;
        $apiLog['text'] = config('errorcode.9999');
        $apiLog['result'] = json_encode([ 'security_key'  => $securityKey, 'expiration'    => $seckey_exp_date]);
        $apiLog['end_time'] = time();
        addApiLog($apiLog);//日志记录
        if($debugging)  ApiSucceds($apiLog);//联调账号联调成功次数统计

        //返回SecurityKey和最后过期时间
        return returnJson([
            'code'  => 0,
            'text'  =>config('errorcode.9999'),
            'result'    => [
                'security_key'  => $securityKey,
                'expiration'    => $seckey_exp_date
            ]
        ]);
    }

    //根据算法参数生成SecurityKey操作
    private function createSecurityKey($algorithm='sha1',$str)
    {
        return hash($algorithm,$str);
    }

}