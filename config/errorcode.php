<?php
/**
 * Created by PhpStorm.
 * User: chengkang
 * Date: 2017/2/22
 * Time: 13:32
 */
return [

    /**
     * 0是返回成功
     * 9开头是通用错误类型
     * 1开头是会员相关
     * 2开头是注单相关
     * 3开头是存取款相关
     *
     */
    9999 => 'Success',//成功
    9001 => 'System maintenance',//系统维护
    9002 => 'Invalid arguments',//请求参数错误
    9003 => 'IP address not allow',//IP限制
    9004 => 'Invalid Agent',//代理错误
    9005 => 'Invalid Agent Status',//代理帐号已停用
    9006 => 'SecurityKey Unable to merge',//SecurityKey获取失败
    9007 => "SecurityKey Has expired",//SecurityKey已经过期
    9008 => "Request is too frequently",//请求太频繁
    9009 => "Data is empty",//数据为空

    1001 => 'Locked',//账户被锁定|停用
    1002 => 'Agent Account Error',//代理帐号已停用
    1003 => 'Account Format Error',//会员帐号格式错误
    1004 => '10 seconds after login',//请10秒后再登录
    1005 => 'Create member error',//创建会员失败
    1006 => 'Exists UserName',//会员帐号已存在
    1007 => 'User not exist',//用户不存在
    1008 => 'Serial not exist',//流水号不存在

    2001 => "The time span is too big",//时间跨度太大

    3001 => 'Insufficient balance',//余额不足
    3002 => 'Withdrawal failed',//取款失败
    3003 => 'Deposit failure',//存款失败


];