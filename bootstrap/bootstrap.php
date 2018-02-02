<?php
/**
 * Created by PhpStorm.
 * User: liangxz@szljfkj.com
 * Date: 2017/3/14
 * Time: 17:27
 */
namespace bootstrap;

use Medoo\Medoo;
use MongoDB\Client;
use bootstrap\MyPDO;

final class Bootstrap
{
    private static $capsule;
    private static $mongodb;
    private static $redis;
    private static $databaseConfig;

    /**
     * Bootstrap constructor.
     * 设置为私有方法，防止外部 new 操作
     */
      private function __construct()
      {

      }  

     /**
     * 获取用户配置信息
     */ 
    public static function getDbConfig()
    {
        if(!empty(self::$databaseConfig)){
            return self::$databaseConfig;
        }

        return require self::configDir().'databases.php';
    }

    /**
     * 获取MYSQLPDO连接对象
     */
    public static function mysqlPDO()
    {
        self::$databaseConfig = self::getDbConfig();
        return MyPDO::getInstance(
            self::$databaseConfig ['mysql']['server'],
            self::$databaseConfig ['mysql']['username'],
            self::$databaseConfig ['mysql']['password'],
            self::$databaseConfig ['mysql']['database_name'],
            self::$databaseConfig ['mysql']['port'],
            self::$databaseConfig ['mysql']['charset']
        );
    }

    /**
     * 获取环境的Config目录
     */
    public static function configDir()
    {
        // 添加配置只需要改动引入配置即可
        switch (SETTING){
            case  'LOCAL':  return  __DIR__.'/../config-local/';
            case  'TEST':   return  __DIR__ . '/../config-test/';
            case  'ONLINE': return __DIR__.'/../config-online/';
            case  'PUBLIC': return  __DIR__.'/../config-public/';
        }
    }
    

    /**
     * 获取MYSQL连接对象
     * 只能通过静态方法进行调用
     */
    public static function DB()
    {

        if(!self::$capsule){
            self::$databaseConfig = self::getDbConfig();
            self::$capsule = new Medoo(self::$databaseConfig['mysql']);
        }

        return self::$capsule;
    }

    /**
     * 获取Redis连接对象
     */
    public static function Redis($database = 'default')
    {
        if( !isset( self::$redis[$database]) || !self::$redis[$database]){
            $redisConfig = self::getDbConfig()['redis'][$database];
            self::$redis[$database] = new \Redis();
            self::$redis[$database]->connect($redisConfig['host'], $redisConfig['port']);
            self::$redis[$database]->auth($redisConfig['password']);
            return  self::$redis[$database];
        }

        return self::$redis[$database];
    }

    /**
     * 获取mongodb连接对象
     */
    public static function MongoDB()
    {
        if(!self::$mongodb){
            self::$databaseConfig = self::getDbConfig();
            self::$mongodb = new Client(self::$databaseConfig['mongodb']['host']);;
        }

        return self::$mongodb;
    }

    /**
     * 获取AMQPConfig
     */
    public static function AMQPConfig()
    {
        self::$databaseConfig = self::getDbConfig();
        return self::$databaseConfig['rabbitmq'];
    }

    /**
     * 防止被克隆
     */
    final function __clone()
    {
        // TODO: Implement __clone() method.
    }


}

