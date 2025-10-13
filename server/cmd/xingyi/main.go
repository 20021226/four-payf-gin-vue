package main

import (
    "fmt"
    "github.com/flipped-aurora/gin-vue-admin/server/utils/pay/xingyi"
    "log"
)

func main() {
    ocr := xingyi.RemoteOCR{URL: "http://192.168.0.208:9580/ocr"}
    svc := xingyi.NewService(nil, ocr)
    // 示例 merId，真实项目中从业务入参传入
    merId := "123456"
    pwd, err := svc.GetLoginResult("15976315080", "CZxiao0768", merId)
    if err != nil {
        log.Fatalf("login error: %v", err)
    }
    fmt.Println("PWD_RESPDATA:", pwd)

    token, err := svc.GetAccessToken(pwd)
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
