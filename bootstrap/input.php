<?php
/**
 * Created by PhpStorm.
 * User: liangxz@szljfkj.com
 * Date: 2017/3/20
 * Time: 13:07
 * 接受input提交参数
 */
namespace bootstrap;

class  Input
{

    /**
     * 获取get参数方法
     * @param null $index
     * @param null $default
     * @return array|null
     */
    public static function get($index = NULL,$default=NULL)
    {
        return self::_fetch_from_array($_REQUEST, $index,$default);
    }

    // --------------------------------------------------------------------

    /**
     * 获取POST参数方法
     * @param null $index
     * @param null $default
     * @return array|null
     */
    public static function post($index = NULL,$default=NULL)
    {
        return self::_fetch_from_array($_REQUEST, $index,$default);
    }


    /**
     * 解析获取$_REQUEST参数
     * @param $array
     * @param null $index
     * @param null $default
     * @return array|null
     */
    public static function _fetch_from_array(&$array, $index = NULL,$default)
    {

        // If $index is NULL, it means that the whole $array is requested
        isset($index) OR $index = array_keys($array);

        // allow fetching multiple keys at once
        if (is_array($index))
        {
            $output = array();
            foreach ($index as $key)
            {
                $output[$key] = self::_fetch_from_array($array, $key,$default);
            }

            return $output;
        }

        if (isset($array[$index]))
        {
            $value = $array[$index];
        }
        elseif (($count = preg_match_all('/(?:^[^\[]+)|\[[^]]*\]/', $index, $matches)) > 1) // Does the index contain array notation
        {
            $value = $array;
            for ($i = 0; $i < $count; $i++)
            {
                $key = trim($matches[0][$i], '[]');
                if ($key === '') // Empty notation will return the value as array
                {
                    break;
                }

                if (isset($value[$key]))
                {
                    $value = $value[$key];
                }
                else
                {
                    return $default;
                }
            }
        }
        else
        {
            return $default;
        }

        return $value;
    }
}