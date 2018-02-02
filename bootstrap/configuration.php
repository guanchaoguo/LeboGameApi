<?php
/**
 * Created by PhpStorm.
 * User: liangxz@szljfkj.com
 * Date: 2017/3/15
 * Time: 15:44
 * 加载配置文件操作
 */
namespace bootstrap;

use bootstrap\Bootstrap;

class Configuration
{
   public static function env($key = "",$file='common')
   {
        if(empty($key))
            return false;

        return self::getName($key, $file);
   }

    /**
     * @param $key
     * @return bool|mixed
     * 配置文件目前只支持common.php，请不要随意修改
     * 前期只支持一维数组的形式获取
     */

   private static function getName($key,$file='common')
   {
       $fileName = Bootstrap::configDir().$file.'.php' ;

       if(!file_exists($fileName))
           return false;

       $commonArr = require $fileName;
       if(!$commonArr || !is_array($commonArr) || !$commonArr[$key])
           return false;

       return $commonArr[$key];
   }
}