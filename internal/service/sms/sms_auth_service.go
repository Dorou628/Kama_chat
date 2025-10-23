package sms

import (
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dypnsapi20170525 "github.com/alibabacloud-go/dypnsapi-20170525/v3/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"kama_chat_server/internal/config"
	"kama_chat_server/pkg/constants"
	"kama_chat_server/pkg/zlog"
)

var smsAuthClient *dypnsapi20170525.Client

// createSmsAuthClient 创建号码认证服务客户端
func createSmsAuthClient() (result *dypnsapi20170525.Client, err error) {
	cfg := config.GetConfig()
	accessKeyID := cfg.SmsAuthConfig.AccessKeyID
	accessKeySecret := cfg.SmsAuthConfig.AccessKeySecret
	
	if smsAuthClient == nil {
		config := &openapi.Config{
			AccessKeyId:     tea.String(accessKeyID),
			AccessKeySecret: tea.String(accessKeySecret),
		}
		// 号码认证服务的endpoint
		config.Endpoint = tea.String("dypnsapi.aliyuncs.com")
		smsAuthClient, err = dypnsapi20170525.NewClient(config)
	}
	return smsAuthClient, err
}

// SendSmsVerifyCode 使用号码认证服务发送短信验证码
func SendSmsVerifyCode(telephone string) (string, int) {
	client, err := createSmsAuthClient()
	if err != nil {
		zlog.Error("创建号码认证服务客户端失败: " + err.Error())
		return constants.SYSTEM_ERROR, -1
	}

	// 构建发送短信验证码请求
	sendSmsRequest := &dypnsapi20170525.SendSmsVerifyCodeRequest{
		PhoneNumber:    tea.String(telephone),
		SignName:       tea.String("速通互联验证码"), // 使用系统赠送签名
		TemplateCode:   tea.String("100001"),      // 使用系统赠送模板
		TemplateParam:  tea.String("{\"code\":\"##code##\",\"min\":\"5\"}"),
		CodeLength:     tea.Int64(6),              // 验证码长度
		ValidTime:      tea.Int64(300),            // 有效时间5分钟
		DuplicatePolicy: tea.Int64(1),             // 覆盖处理
		Interval:       tea.Int64(60),             // 间隔60秒
		CodeType:       tea.Int64(1),              // 纯数字
		ReturnVerifyCode: tea.Bool(true),          // 返回验证码
	}

	runtime := &util.RuntimeOptions{}
	
	// 调用阿里云号码认证服务发送短信验证码
	rsp, err := client.SendSmsVerifyCodeWithOptions(sendSmsRequest, runtime)
	if err != nil {
		zlog.Error("发送短信验证码失败: " + err.Error())
		return constants.SYSTEM_ERROR, -1
	}

	// 检查响应结果
	if rsp.Body != nil && rsp.Body.Code != nil && *rsp.Body.Code == "OK" {
		zlog.Info("短信验证码发送成功: " + *util.ToJSONString(rsp))
		return "验证码发送成功，请及时查收短信", 0
	} else {
		errorMsg := "短信发送失败"
		if rsp.Body != nil && rsp.Body.Message != nil {
			errorMsg = *rsp.Body.Message
		}
		zlog.Error("短信发送失败: " + errorMsg)
		return errorMsg, -1
	}
}

// CheckSmsVerifyCode 验证短信验证码
func CheckSmsVerifyCode(telephone, inputCode string) (bool, string) {
	client, err := createSmsAuthClient()
	if err != nil {
		zlog.Error("创建号码认证服务客户端失败: " + err.Error())
		return false, constants.SYSTEM_ERROR
	}

	// 构建验证请求
	checkSmsVerifyCodeRequest := &dypnsapi20170525.CheckSmsVerifyCodeRequest{
		PhoneNumber: tea.String(telephone),
		VerifyCode:  tea.String(inputCode),
		OutId:       tea.String(""),
	}

	runtime := &util.RuntimeOptions{}
	
	// 直接调用阿里云号码认证服务验证验证码
	rsp, err := client.CheckSmsVerifyCodeWithOptions(checkSmsVerifyCodeRequest, runtime)
	if err != nil {
		zlog.Error("验证码校验失败: " + err.Error())
		return false, constants.SYSTEM_ERROR
	}

	// 检查验证结果
	if rsp.Body != nil && rsp.Body.Code != nil && *rsp.Body.Code == "OK" {
		zlog.Info("验证码校验成功: " + *util.ToJSONString(rsp))
		return true, "验证码校验成功"
	} else {
		errorMsg := "验证码校验失败"
		if rsp.Body != nil && rsp.Body.Message != nil {
			errorMsg = *rsp.Body.Message
		}
		zlog.Error("验证码校验失败: " + errorMsg)
		return false, errorMsg
	}
}