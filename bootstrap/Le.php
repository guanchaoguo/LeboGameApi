<?php
/**
 * Created by PhpStorm.
 * User: liangxz@szljfkj.com
 * Date: 2017/3/20
 * Time: 14:27
 */

namespace bootstrap;

class Le_framework
{
    public static function boot()
    {
        $dir = __DIR__."/";
        if(is_dir($dir))
        {
            $files = scandir($dir);
            foreach ($files as $val)
            {
                $fileArr = explode(".",$val);
                if($fileArr[1] != 'php' || $fileArr[0] == 'Le')
                    continue;
                require $dir.$val;
            }
        }
        require __DIR__.'/../vendor/autoload.php';
    }
}