package main

import (
	"fmt"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/pay/xianxiang"
	"go.uber.org/zap"
	"log"
	"time"
)

// initLogger 初始化基本的日志系统
func initLogger() {
	// 创建一个开发环境的 zap logger 配置
	config := zap.NewDevelopmentConfig()
	config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)

	// 创建 logger
	logger, err := config.Build()
	if err != nil {
		log.Fatalf("初始化日志系统失败: %v", err)
	}

	// 设置全局 logger
	global.GVA_LOG = logger

	global.GVA_LOG.Info("日志系统初始化完成")
}

func main() {
	// 初始化基本的日志系统
	initLogger()

	defer func() {
		if global.GVA_LOG != nil {
			global.GVA_LOG.Sync()
		}
	}()

	global.GVA_LOG.Info("开始执行 xianxiang 测试")

	hash := utils.BcryptHash("123456")
	global.GVA_LOG.Info(hash)

	// 创建 xianxiang 服务实例
	svc := xianxiang.NewService(nil)

	// 示例用户名和密码，真实项目中从配置或参数传入
	username := "9902911"
	password := "123456"

	// 1. 获取访问令牌
	loginResult, err := svc.GetToken(username, password)
	if err != nil {
		log.Fatalf("获取访问令牌失败: %v", err)
	}
	fmt.Println("ACCESS_TOKEN:", loginResult.AccessToken)
	fmt.Println("TOKEN_TYPE:", loginResult.TokenType)
	fmt.Println("EXPIRES_IN:", loginResult.ExpiresIn)

	// 3. 获取订单列表
	// 使用今天的日期作为查询时间
	today := time.Now().Format("2006-01-02")
	fmt.Printf("查询日期: %s\n", today)

	body, err := svc.GetOrderList(loginResult.AccessToken, today)
	if err != nil {
		log.Fatalf("获取订单列表失败: %v", err)
	}
	fmt.Println("订单列表响应:")
	fmt.Println(string(body))

	// 4. 测试支付检查功能
	// 设置测试时间范围（今天的开始和结束）
	startTime := time.Now().Truncate(24 * time.Hour)     // 今天 00:00:00
	endTime := startTime.Add(24*time.Hour - time.Second) // 今天 23:59:59
	testAmount := "100.00"                               // 测试金额

	fmt.Printf("测试支付检查 - 金额: %s, 时间范围: %s 到 %s\n",
		testAmount,
		startTime.Format("2006-01-02 15:04:05"),
		endTime.Format("2006-01-02 15:04:05"))

	found, orderInfo, err := svc.CheckPayment(string(body), startTime, endTime, testAmount)
	if err != nil {
		log.Fatalf("检查支付结果失败: %v", err)
	}

	if found {
		fmt.Printf("✅ 找到匹配的支付订单 (金额: %s)\n", testAmount)
		if orderInfo != nil {
			fmt.Printf("   订单创建时间: %s\n", orderInfo.CreateAt)
			fmt.Printf("   订单号: %s\n", orderInfo.OrderSn)
			fmt.Printf("   实收金额: %s\n", orderInfo.RealAmount)
		}
	} else {
		fmt.Printf("❌ 未找到匹配的支付订单 (金额: %s)\n", testAmount)
	}

	// 5. 测试令牌验证
	fmt.Println("测试令牌验证...")
	isValid := svc.CheckToken(loginResult.AccessToken)
	if isValid {
		fmt.Println("✅ 令牌验证通过")
	} else {
		fmt.Println("❌ 令牌验证失败")
	}

	fmt.Println("xianxiang 测试完成")
}
