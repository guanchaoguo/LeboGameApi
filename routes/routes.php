<?php
/**
 * Created by PhpStorm.
 * User: liangxz@szljfkj.com
 * Date: 2017/3/14
 * Time: 13:41
 * 框架路由文件
 * 路由规则都写在这个文件中
 */
use NoahBuscher\Macaw\Macaw;

Macaw::get('/mycat','App\controllers\mycat@mycat');
Macaw::get('/',function(){
        echo "Welcome to visit Lebo open platform!\n
歡迎訪問利博遊戲開放平台";
});
Macaw::get('/doc','App\controllers\SwaggerController@doc');
//接入商获取SecurityKey
Macaw::post('/getKey','App\controllers\KeyController@createKey');

//会员充值操作
Macaw::post('/deposit','App\controllers\CashController@deposit');

//注单查询
Macaw::post('/orderList','App\controllers\UserCashController@orderList');

//根据时间段获取注单信息
Macaw::post('/getDateList','App\controllers\UserCashController@getOrderListByDate');

/**
 * 会员登录&注册
 */
Macaw::post('/authorization','App\controllers\AuthController@login');

/**
 * 获取供应商会员信息
 */
Macaw::post('user', 'App\controllers\PlayerController@index');

/**
 * 会员取款
 */
Macaw::post('withDrawal', 'App\controllers\PlayerController@withDrawal');

/**
 * 会员存取款状态查询
 */
Macaw::post('transferLog', 'App\controllers\TransferLogController@index');

/**

 * 发送消息
 */
Macaw::get('send', 'App\controllers\MqProducerController@publishMsg');


/**
 * 接收消息
 */
Macaw::get('rev', 'App\controllers\ReceiveController@getMsg');

/*
 * 根据时间段获取异常注单信息
 */
Macaw::post('exceptionOrderLog', 'App\controllers\ExceptionOrderLogController@index');

Macaw::dispatch();