<?php

namespace App\controllers;

class SwaggerController
{

    public function __construct()
    {
    }

    /**
     * 返回JSON格式的Swagger定义
     * 这里需要一个主`Swagger`定义：
     * @SWG\Swagger(
     *   schemes={"http"},
     *   @SWG\Info(
     *     title="LEBO游戏API接口文档",
     *     version="1.0.0",
     *     description="
     *     本文档主要叙述LEBO游戏中心服务器与游戏接入商服务器之间的数据通讯协议及规范，
     *     双方的软件在此基础上达到数据高性能、高安全性进行交换与共享的目的。
     *     code状态码：
          [
            0    : 'Success',//成功
            9001 : 'System maintenance',//系统维护
            9002 : 'Invalid arguments',//请求参数错误
            9003 : 'IP address not allow',//IP限制
            9004 : 'Invalid Agent',//代理错误
            9005 : 'Invalid Agent Status',//代理帐号已停用
            9006 : 'SecurityKey Unable to merge',//SecurityKey获取失败
            9007 : 'SecurityKey Has expired',//SecurityKey已经过期
            9008 : 'Request is too frequently',//请求太频繁
            9009 : 'Data is empty',//数据为空
            1001 : 'Locked',//账户被锁定|停用
            1002 : 'Agent Account Error',//代理帐号已停用
            1003 : 'Account Format Error',//会员帐号格式错误
            1004 : '10 seconds after login',//请10秒后再登录
            1005 : 'Create member error',//创建会员失败
            1006 : 'Exists UserName',//会员帐号已存在
            1007 : 'User not exist',//用户不存在
            1008 : 'Serial not exist',//流水号不存在
            2001 : 'The time span is too big',//时间跨度太大
            3001 : 'Insufficient balance',//余额不足
            3002 : 'Withdrawal failed',//取款失败
            3003 : 'Deposit failure',//存款失败
          ]"
     *   ),
     * )
     */
    public function doc()
    {
        // 你可以将API的`Swagger Annotation`写在实现API的代码旁，从而方便维护，
        // `swagger-php`会扫描你定义的目录，自动合并所有定义。这里我们直接用`Controller/`
        // 文件夹。
        $swagger = \Swagger\scan(__DIR__);
        die(json_encode($swagger));
    }
}
