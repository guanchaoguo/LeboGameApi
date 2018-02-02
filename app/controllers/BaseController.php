<?php

namespace App\controllers;

class BaseController
{

    function  __construct()
    {
        IpLimit::init();
    }





    /**
     * 根据agent获取对应的SecurityKey
     * @param $agent
     * @return bool
     */
    protected function getSecurityKey($agent)
    {
        if(empty($agent))
            return false;

        //先从redis中获取，如果没有再进行数据库查询
        $key_name = $agent.env('AGENT_IPS');
        $result =json_decode(Redis::get($key_name));

        if(!$result)
        {
            $agentObj = Agent::where('user_name', $agent)->first();
            $result = $agentObj->getWhites;
        }

        if(!$result)
            return false;

        return $result->agent_seckey;
    }

    /**
     * 返回数据验证错误代码
     * @return mixed
     */
    protected function returnError()
    {
        return returnJson([
            'code'  => 9002,
            'text'  => config('errorcode.9002'),
            'result'    => ''
        ]);
    }

}
