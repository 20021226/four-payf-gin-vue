package example

import (
	"context"
	"fmt"
	"github.com/shopspring/decimal"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/example"
	exampleReq "github.com/flipped-aurora/gin-vue-admin/server/model/example/request"
	exampleRes "github.com/flipped-aurora/gin-vue-admin/server/model/example/response"
	payReq "github.com/flipped-aurora/gin-vue-admin/server/model/pay/request"
	payTask "github.com/flipped-aurora/gin-vue-admin/server/task"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type MerUserApi struct{}

// checkAmountRangeConflict 检查指定金额范围内是否存在冲突的 Redis 键
// 输入金额如 5.00，检查 5.01-5.99 范围内是否有键存在
// 返回 true 表示存在冲突，false 表示可以使用
func (merUserApi *MerUserApi) checkAmountRangeConflict(ctx context.Context, userID uint, inputAmount int64) (bool, error) {
	// 计算检查范围：输入金额的下一分到下一元的前一分
	// 例如：输入 5.00，检查 5.01 到 5.99
	baseAmount := float64(inputAmount)
	startAmount := math.Floor(baseAmount*100+1) / 100 // 5.01
	endAmount := math.Floor(baseAmount) + 0.99        // 5.99

	// 构建 Redis 键的模式
	keyPattern := fmt.Sprintf("%s:%d:*", global.PAY_AMOUNT_USED_KEY, userID)

	// 使用 SCAN 命令遍历匹配的键
	iter := global.GVA_REDIS.Scan(ctx, 0, keyPattern, 100).Iterator()

	for iter.Next(ctx) {
		key := iter.Val()

		// 从键中提取金额部分
		parts := strings.Split(key, ":")
		if len(parts) < 3 {
			continue
		}

		amountStr := parts[2]
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			continue
		}

		// 检查是否在冲突范围内
		if amount >= startAmount && amount <= endAmount {
			return true, nil // 发现冲突
		}
	}

	if err := iter.Err(); err != nil {
		return false, fmt.Errorf("Redis SCAN 操作失败: %w", err)
	}

	return false, nil // 无冲突
}

// findAvailableAmount 查找指定金额范围内未被占用的金额
// inputAmount: 输入的基础金额（单位: 元）
// minDecimalAmount, maxDecimalAmount: 小数部分范围（如 1-99 表示 0.01-0.99）
// 例如：输入 500（5.00元），在 5.01-5.99 范围内查找未被占用的金额
// 返回可用的金额，如果全部被占用则返回 0
func (merUserApi *MerUserApi) findAvailableAmount(ctx context.Context, userID uint, inputAmount int64, maxDecimalAmount, minDecimalAmount int32) (decimal.Decimal, error) {
	// 将输入金额转换为正确的浮点数格式
	// inputAmount 单位: 元
	baseAmount := decimal.NewFromInt(inputAmount)

	// 构建 Redis 键的模式
	keyPattern := fmt.Sprintf("%s:%d:*", global.PAY_AMOUNT_USED_KEY, userID)

	// 获取所有已占用的金额
	occupiedAmounts := make(map[string]bool)
	iter := global.GVA_REDIS.Scan(ctx, 0, keyPattern, 100).Iterator()

	for iter.Next(ctx) {
		key := iter.Val()

		// 从键中提取金额部分
		parts := strings.Split(key, ":")
		if len(parts) < 3 {
			continue
		}

		amountStr := parts[2]
		occupiedAmounts[amountStr] = true
	}

	if err := iter.Err(); err != nil {
		return decimal.Zero, fmt.Errorf("Redis SCAN 操作失败: %w", err)
	}

	// 收集所有可用的金额
	var availableAmounts []decimal.Decimal

	// 在指定的小数范围内查找所有未被占用的金额
	// 例如：baseAmount = 5.00, minDecimalAmount = 1, maxDecimalAmount = 99
	// 会检查 5.01, 5.02, ..., 5.99
	for cent := minDecimalAmount; cent <= maxDecimalAmount; cent++ {
		candidateAmount := baseAmount.Add(
			decimal.NewFromInt(int64(cent)).Div(decimal.NewFromInt(100)),
		)
		candidateAmountStr := candidateAmount.StringFixed(2)

		if !occupiedAmounts[candidateAmountStr] {
			availableAmounts = append(availableAmounts, candidateAmount)
		}
	}

	// 如果当前范围有可用金额，随机选择一个
	if len(availableAmounts) > 0 {
		randomIndex := rand.Intn(len(availableAmounts))
		return availableAmounts[randomIndex], nil
	}

	return decimal.Zero, nil // 全部被占用
}

// checkAmountRangeConflictOptimized 优化版本：使用更精确的范围检查
func (merUserApi *MerUserApi) checkAmountRangeConflictOptimized(ctx context.Context, userID uint, inputCents int64) (bool, error) {
	// 将金额转换为分（避免浮点数精度问题）

	// 计算检查范围（分为单位）
	startCents := inputCents + 1
	endCents := (inputCents/100)*100 + 99 // 同一元内的最大值

	// 如果输入已经是 x.99，则检查下一元的范围
	if inputCents%100 == 99 {
		startCents = inputCents + 2 // 跳到下一元的 .01
		endCents = startCents + 98  // 到下一元的 .99
	}

	keyPattern := fmt.Sprintf("%s:%d:*", global.PAY_AMOUNT_USED_KEY, userID)

	// 使用批量检查，减少网络往返
	var cursor uint64
	var conflictFound bool

	for {
		keys, nextCursor, err := global.GVA_REDIS.Scan(ctx, cursor, keyPattern, 50).Result()
		if err != nil {
			return false, fmt.Errorf("Redis SCAN 失败: %w", err)
		}

		for _, key := range keys {
			parts := strings.Split(key, ":")
			if len(parts) < 3 {
				continue
			}

			amountStr := parts[2]
			amount, err := strconv.ParseFloat(amountStr, 64)
			if err != nil {
				continue
			}

			amountCents := int64(math.Round(amount * 100))

			// 检查是否在冲突范围内
			if amountCents >= startCents && amountCents <= endCents {
				conflictFound = true
				break
			}
		}

		if conflictFound || nextCursor == 0 {
			break
		}
		cursor = nextCursor
	}

	return conflictFound, nil
}

// CreateMerUser 创建merUser表
// @Tags MerUser
// @Summary 创建merUser表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body example.MerUser true "创建merUser表"
// @Success 200 {object} response.Response{msg=string} "创建成功"
// @Router /merUser/createMerUser [post]
func (merUserApi *MerUserApi) CreateMerUser(c *gin.Context) {
	// 创建业务用Context
	ctx := c.Request.Context()

	var merUser example.MerUser
	err := c.ShouldBindJSON(&merUser)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = utils.Verify(merUser, utils.MerUserVerify)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = merUserService.CreateMerUser(ctx, &merUser)
	if err != nil {
		global.GVA_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// DeleteMerUser 删除merUser表
// @Tags MerUser
// @Summary 删除merUser表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body example.MerUser true "删除merUser表"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /merUser/deleteMerUser [delete]
func (merUserApi *MerUserApi) DeleteMerUser(c *gin.Context) {
	// 创建业务用Context
	ctx := c.Request.Context()

	id := c.Query("id")
	err := merUserService.DeleteMerUser(ctx, id)
	if err != nil {
		global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// DeleteMerUserByIds 批量删除merUser表
// @Tags MerUser
// @Summary 批量删除merUser表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{msg=string} "批量删除成功"
// @Router /merUser/deleteMerUserByIds [delete]
func (merUserApi *MerUserApi) DeleteMerUserByIds(c *gin.Context) {
	// 创建业务用Context
	ctx := c.Request.Context()

	ids := c.QueryArray("ids[]")
	err := merUserService.DeleteMerUserByIds(ctx, ids)
	if err != nil {
		global.GVA_LOG.Error("批量删除失败!", zap.Error(err))
		response.FailWithMessage("批量删除失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("批量删除成功", c)
}

// UpdateMerUser 更新merUser表
// @Tags MerUser
// @Summary 更新merUser表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body example.MerUser true "更新merUser表"
// @Success 200 {object} response.Response{msg=string} "更新成功"
// @Router /merUser/updateMerUser [put]
func (merUserApi *MerUserApi) UpdateMerUser(c *gin.Context) {
	// 从ctx获取标准context进行业务行为
	ctx := c.Request.Context()

	var merUser example.MerUser
	err := c.ShouldBindJSON(&merUser)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = utils.Verify(merUser, utils.MerUserVerify)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = merUserService.UpdateMerUser(ctx, merUser)
	if err != nil {
		global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// FindMerUser 用id查询merUser表
// @Tags MerUser
// @Summary 用id查询merUser表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param id query int true "用id查询merUser表"
// @Success 200 {object} response.Response{data=example.MerUser,msg=string} "查询成功"
// @Router /merUser/findMerUser [get]
func (merUserApi *MerUserApi) FindMerUser(c *gin.Context) {
	// 创建业务用Context
	ctx := c.Request.Context()

	id := c.Query("id")
	remerUser, err := merUserService.GetMerUser(ctx, id)
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败:"+err.Error(), c)
		return
	}
	response.OkWithData(remerUser, c)
}

// GetMerUserList 分页获取merUser表列表
// @Tags MerUser
// @Summary 分页获取merUser表列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query exampleReq.MerUserSearch true "分页获取merUser表列表"
// @Success 200 {object} response.Response{data=response.PageResult,msg=string} "获取成功"
// @Router /merUser/getMerUserList [get]
func (merUserApi *MerUserApi) GetMerUserList(c *gin.Context) {
	// 创建业务用Context
	ctx := c.Request.Context()
	var pageInfo exampleReq.MerUserSearch
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := merUserService.GetMerUserInfoList(ctx, pageInfo)
	if err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     pageInfo.Page,
		PageSize: pageInfo.PageSize,
	}, "获取成功", c)
}

// GetMerUserPublic 不需要鉴权的merUser表接口
// @Tags MerUser
// @Summary 不需要鉴权的merUser表接口
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /merUser/getMerUserPublic [get]
func (merUserApi *MerUserApi) GetMerUserPublic(c *gin.Context) {
	// 创建业务用Context
	ctx := c.Request.Context()

	// 此接口不需要鉴权
	// 示例为返回了一个固定的消息接口，一般本接口用于C端服务，需要自己实现业务逻辑
	merUserService.GetMerUserPublic(ctx)
	response.OkWithDetailed(gin.H{
		"info": "不需要鉴权的merUser表接口信息",
	}, "获取成功", c)
}

// CreateMerUser 创建merUser表
// @Tags MerUser
// @Summary 创建merUser表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body example.MerUser true "创建merUser表"
// @Success 200 {object} response.Response{msg=string} "创建成功"
// @Router /merUser/createMerUser [post]
func (merUserApi *MerUserApi) GetPayQrCode(c *gin.Context) {
	// 创建业务用Context
	ctx := c.Request.Context()
	// 支持永久token和JWT两种方式获取用户ID
	userID := middleware.GetUserIDFromTokenOrJWT(c)
	var reqParms payReq.PayQrcodeParms
	var paymentQrCodeResponse exampleRes.PaymentQrCodeResponse
	err := c.ShouldBindJSON(&reqParms)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	// 参数验证
	if err := utils.Verify(reqParms, utils.PayQrcodeParmsVerify); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	//get user config
	sysUserConfig, err := sysUserConfigService.GetConfigBySysUserID(ctx, int64(userID))
	if err != nil {
		global.GVA_LOG.Error("获取用户配置失败!", zap.Error(err))
		response.StdFail(c, "获取用户配置失败:"+err.Error())
		return
	}
	// 校验允许的域名/IP
	if !utils.IsHostAllowed(sysUserConfig.AllowRequestUrl, c.Request.Host) {
		response.StdFail(c, "域名或IP未授权调用")
		return
	}

	//get mer data
	merUserList, err := merUserService.GetNomalMerUser(ctx, reqParms, userID)
	if err != nil {
		global.GVA_LOG.Error("获取普通商户用户失败!", zap.Error(err))
		response.StdFail(c, "获取普通商户用户失败:"+err.Error())
		return
	}

	currentMerUser := example.MerUser{}
	if reqParms.MerId != nil {
		for _, item := range merUserList {
			if item.Id != nil && reqParms.MerId != nil && *item.Id == *reqParms.MerId {
				currentMerUser = item
				break
			}
		}
	} else {
		// 随机从列表中选择一个
		if len(merUserList) > 0 {
			randomIndex := rand.Intn(len(merUserList))
			currentMerUser = merUserList[randomIndex]
		}
	}

	if currentMerUser.Id == nil {
		response.StdFail(c, "未找到商户的收款码,请登录后台查看商户状态")
		return
	}

	if currentMerUser.MaxAmount == nil || currentMerUser.MinAmount == nil {
		response.StdFail(c, "商户请求金额范围未配置")
		return
	}
	if reqParms.PayAmmount > *currentMerUser.MaxAmount || reqParms.PayAmmount < *currentMerUser.MinAmount {
		response.StdFail(c, "金额不符合设置的范围")
		return
	}

	//// 检查金额范围冲突：防止相近金额的重复支付
	//hasConflict, err := merUserApi.checkAmountRangeConflictOptimized(ctx, userID, reqParms.PayAmmount)
	//if err != nil {
	//	global.GVA_LOG.Error("检查金额冲突失败!", zap.Error(err))
	//	response.StdFail(c, "该金额已被占用，请稍后重试")
	//	return
	//}

	// 处理可能的 nil 指针
	var maxDecimalAmount int32
	var minDecimalAmount int32

	if currentMerUser.MaxDecimalAmount != nil {
		maxDecimalAmount = *currentMerUser.MaxDecimalAmount
	} else {
		// 设置默认值或处理错误
		maxDecimalAmount = 100 // 默认值
	}

	if currentMerUser.MinDecimalAmount != nil {
		minDecimalAmount = *currentMerUser.MinDecimalAmount
	} else {
		minDecimalAmount = 1 // 默认值
	}

	availableAmount, err := merUserApi.findAvailableAmount(ctx, userID, reqParms.PayAmmount, maxDecimalAmount, minDecimalAmount)
	if err != nil {
		global.GVA_LOG.Error("查找可用金额失败!", zap.Error(err))
		response.StdFail(c, "系统繁忙，请稍后重试")
		return
	}

	if availableAmount == decimal.Zero {
		response.StdFail(c, fmt.Sprintf("金额 %.2f 及其相近金额范围内的所有金额都已被占用，请稍后重试或使用其他金额", float64(reqParms.PayAmmount)))
		return
	}

	// 使用找到的可用金额
	paymentQrCodeResponse.Amount = &availableAmount
	global.GVA_LOG.Info(
		"自动调整金额",
		zap.Float64("原金额", float64(reqParms.PayAmmount)),
		zap.String("调整后金额", availableAmount.String()),
	)

	currentMerUser = merUserList[rand.Intn(len(merUserList))]
	currentTimeStr := utils.GetCurrentTimeStr()
	paymentQrCodeResponse.QrcodeCode = currentMerUser.QrCode
	paymentQrCodeResponse.CreateTime = &currentTimeStr

	// 生成订单ID
	paymentQrCodeResponse.OrderId = &reqParms.OrderId

	//payCreateTime := utils.GetCurrentTimeStr()

	amount := decimal.NewFromInt(reqParms.PayAmmount)
	payOrder := example.MerPayOrder{
		MerId:          currentMerUser.Id,
		MerName:        currentMerUser.MerName,
		MerType:        currentMerUser.MerType,
		SysUserId:      currentMerUser.SysUserId,
		Ammount:        &amount,
		RequestAmmount: paymentQrCodeResponse.Amount,
		OrderId:        paymentQrCodeResponse.OrderId,
	}
	err = merPayOrderService.CreateMerPayOrder(ctx, &payOrder)
	if err != nil {
		global.GVA_LOG.Error("创建支付订单失败!", zap.Error(err))
		response.StdFail(c, "创建支付订单失败:"+err.Error())
		return
	}

	// 插入成功后，payOrder.Id 会被自动填充为数据库生成的 ID
	global.GVA_LOG.Info("支付订单创建成功",
		zap.Int64("订单ID", *payOrder.Id),
		zap.String("订单号", *payOrder.OrderId))

	// 将订单上下文写入 Redis，便于后续查询与校验
	if paymentQrCodeResponse.OrderId != nil {
		// 确保 Amount 不为空
		amountStr := availableAmount.String()

		var ttl time.Duration
		if reqParms.Expires == 0 {
			ttl = time.Duration(5) * time.Minute
		} else {
			ttl = time.Duration(reqParms.Expires) * time.Second
		}
		startTime, endTime := utils.GetTimeRange(ttl)

		// 构建 Redis 键：pay_amount_used:userID:amount
		redisKey := fmt.Sprintf("%s:%d:%s", global.PAY_AMOUNT_USED_KEY, userID, amountStr)
		ok, err := global.GVA_REDIS.SetNX(ctx, redisKey, 1, ttl).Result()
		if err != nil {
			global.GVA_LOG.Error("缓存订单上下文失败!", zap.Error(err))
			response.StdFail(c, "系统繁忙，请稍后重试")
			return
		} else if !ok {
			response.StdFail(c, fmt.Sprintf("金额 %s 及其相近金额范围内的所有金额都已被占用，请稍后重试或使用其他金额", amountStr))
			return
		} else {
			// Redis SetNX 成功，创建并启动订单监控任务
			monitorTask := payTask.NewOrderMonitorTask(
				userID,
				amountStr,
				redisKey,
				ttl,
				*currentMerUser.MerType,
				*paymentQrCodeResponse.OrderId,
				*currentMerUser.Id,
				startTime,
				endTime,
				reqParms.CallbackUrl,
				*payOrder.Id, // 使用插入后获取的订单 ID
			)
			monitorTask.Start()

			global.GVA_LOG.Info("订单创建成功，监控任务已启动",
				zap.Uint("userID", userID),
				zap.String("amount", amountStr),
				zap.String("redisKey", redisKey),
				zap.Duration("ttl", ttl))
		}
	}
	paymentQrCodeResponse.UniqueId = payOrder.Id
	response.StdOk(c, paymentQrCodeResponse, "创建成功")
}

// GeneratePermanentToken 生成永久token（每个用户只能有一个有效token，生成新token会删除旧token）
// @Tags MerUser
// @Summary 生成永久token（单一token机制）
// @Description 为用户生成永久token，每个用户只能拥有一个有效的永久token。生成新token时会自动删除所有旧token。
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=object,msg=string} "生成成功，旧token已删除"
// @Router /merUser/generatePermanentToken [post]
func (merUserApi *MerUserApi) GeneratePermanentToken(c *gin.Context) {
	userID := utils.GetUserID(c)

	// 生成永久token（会自动删除旧token）
	token, err := utils.GeneratePermanentToken(userID)
	if err != nil {
		global.GVA_LOG.Error("生成永久token失败!", zap.Error(err))
		response.FailWithMessage("生成永久token失败:"+err.Error(), c)
		return
	}

	response.OkWithDetailed(gin.H{
		"token":      token,
		"user_id":    userID,
		"created_at": time.Now().Unix(),
		"notice":     "新token已生成，所有旧token已自动删除",
	}, "生成成功，旧token已删除", c)
}

// GetPermanentTokens 获取用户的永久token列表
// @Tags MerUser
// @Summary 获取用户的永久token列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /merUser/getPermanentTokens [get]
func (merUserApi *MerUserApi) GetPermanentTokens(c *gin.Context) {
	userID := utils.GetUserID(c)

	// 获取用户的所有永久token
	tokens, err := utils.GetUserPermanentTokens(userID)
	if err != nil {
		global.GVA_LOG.Error("获取永久token列表失败!", zap.Error(err))
		response.FailWithMessage("获取永久token列表失败:"+err.Error(), c)
		return
	}

	response.OkWithDetailed(gin.H{
		"tokens": tokens,
		"total":  len(tokens),
	}, "获取成功", c)
}

// RevokePermanentToken 删除永久token
// @Tags MerUser
// @Summary 删除永久token
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body object{token=string} true "删除永久token"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /merUser/revokePermanentToken [post]
func (merUserApi *MerUserApi) RevokePermanentToken(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}

	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	// 删除永久token
	err = utils.RevokePermanentToken(req.Token)
	if err != nil {
		global.GVA_LOG.Error("删除永久token失败!", zap.Error(err))
		response.FailWithMessage("删除永久token失败:"+err.Error(), c)
		return
	}

	response.OkWithMessage("删除成功", c)
}
