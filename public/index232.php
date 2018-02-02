<?php
/**
 * Created by PhpStorm.
 * User: liangxz@szljfkj.com
 * Date: 2017/3/14
 * Time: 13:40
 * 框架入口文件
 */
//自动加载
error_reporting(E_ALL & ~E_NOTICE);
ini_set("display_errors", 1);
define('SETTING','TEST');//环境设定，测试环境为 TEST，线上环境为 ONLINE, 开发环境 LOCAL  外网测试环境 PUBLIC 系统会根据这个常量加载对应环境的配置文件
date_default_timezone_set('Brazil/West');
header("Access-Control-Allow-Origin: *");
require __DIR__.'/../bootstrap/Le.php';
\bootstrap\Le_framework::boot();


