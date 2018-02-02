<?php

namespace App\Model;

/**
 * php内置数据的过滤函数的简单封装用于校验数据合法性
 * 数据校验类
 * Class Validator
 * @package App\Model
 */
class Validator
{
    private $msg = null;

    /**
     * 参数校验
     * @param  array $aParamArr  需要校验的参数
     * @param  array $aFilterArr  校验规则
     * @return boolean  是否校验通过
     *
     * 参考用例：
     *       $_POST['class'] ='w'; // 用户的输入值
     *       $aParamArr = $_POST;  //传入的校验数组 如用户的输入
     *       // 校验规则
     *       $aFilterArr = array(
     *           "class" => array(//需要校验的字段
     *               "required"=>1, //是否为必填参数
     *               "filter"=>FILTER_VALIDATE_INT,//需要校验的参数类型
     *                  // 提示：参见完整的 PHP Filter 参考手册，查看可与该函数一同使用的过滤器。
     *               "options"=>array(
     *                   "min_range" =>0,//最大值
     *                   "max_range" =>6,//最小值
     *                   'default' => '格式非法！',// 错误消息提示类型
     *           )
     *           ),
     *       );
     * 具体调用参考方法：
     *   $validator = new validator;
     *   if(!$validator->make($aParamArr,$aFilterArr)){
     *        var_dump($validator->msg);
     *   }
     */
    public function make($aParamArr,$aFilterArr)
    {
        // 校验结果
        $res = filter_var_array($aParamArr, $aFilterArr);

        foreach ($res as $key => $val) {
            if (is_null($val)) {
                if (!$aFilterArr[$key]['required']) {
                    if (!isset($aFilterArr[$key]['options']['default']))
                        $this->msg =  "参数{$key}为必填！";
                    return false;
                } else {
                    $this->msg =   "缺少参数{$key}！";
                    return  false;
                }
            }
        }

        if (isset($aFilterArr[$key]['options']['default'])) {
            $this->msg = "参数{$key}错误：{$res[$key]}" ;
            return  false;
        }

        return true;
    }

}
