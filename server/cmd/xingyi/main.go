package main

import (
	"fmt"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/pay/xingyi"
	"go.uber.org/zap"
	"log"
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

	global.GVA_LOG.Info("开始执行 xingyi 测试")
	ocr := xingyi.NewXingyiOCR(xingyi.Config{
		BaseURL: "http://192.168.0.208:9580/ocr",
	})
	svc := xingyi.NewService(nil, ocr, xingyi.Cookies{})
	// 示例 merId，真实项目中从业务入参传入
	merId := "123456"
	pwd, err := svc.GetLoginResult("15976315080", "CZxiao0768", merId)
	if err != nil {
		log.Fatalf("login error: %v", err)
	}
	fmt.Println("PWD_RESPDATA:", pwd)
	token, err := svc.GetAccessToken(pwd, merId)
	if err != nil {
		log.Fatalf("get token error: %v", err)
	}
	fmt.Println("ACCESS_TOKEN:", token)

	body, err := svc.GetPayList(token, "20230801000000", "20230801235959", merId)
	if err != nil {
		log.Fatalf("get pay list error: %v", err)
	}
	fmt.Println(string(body))
}
