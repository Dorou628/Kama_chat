package main

import (
	"fmt"
	"kama_chat_server/internal/config"
	"kama_chat_server/internal/dao"
	"kama_chat_server/internal/service/redis"
	"kama_chat_server/internal/service/sms"
	"kama_chat_server/pkg/zlog"
	"time"
)

func main() {
	// 初始化配置
	config.LoadConfig()
	
	// 初始化数据库连接
	dao.InitGorm()
	
	// 初始化Redis连接
	redis.InitRedis()
	
	// 初始化日志
	zlog.InitLogger()
	
	fmt.Println("=== 阿里云号码认证服务测试 ===")
	
	// 测试手机号（请替换为真实的测试手机号）
	testPhone := "13800000000" // 请替换为您的测试手机号
	
	fmt.Printf("测试手机号: %s\n", testPhone)
	
	// 测试发送验证码
	fmt.Println("\n1. 测试发送验证码...")
	message, code := sms.VerificationCode(testPhone)
	fmt.Printf("发送结果: %s (状态码: %d)\n", message, code)
	
	if code == 0 {
		fmt.Println("验证码发送成功！")
		
		// 等待用户输入验证码进行测试
		fmt.Print("\n请输入收到的验证码进行测试: ")
		var inputCode string
		fmt.Scanln(&inputCode)
		
		// 测试验证码校验
		fmt.Println("\n2. 测试验证码校验...")
		isValid, verifyMessage := sms.VerifyCode(testPhone, inputCode)
		fmt.Printf("校验结果: %s (是否有效: %t)\n", verifyMessage, isValid)
		
		if isValid {
			fmt.Println("✅ 验证码校验成功！")
		} else {
			fmt.Println("❌ 验证码校验失败！")
		}
	} else {
		fmt.Printf("❌ 验证码发送失败，错误码: %d\n", code)
	}
	
	fmt.Println("\n=== 测试完成 ===")
}