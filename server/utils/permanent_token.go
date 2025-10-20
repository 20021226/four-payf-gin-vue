package utils

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"go.uber.org/zap"
)

// PermanentToken 永久token结构
type PermanentToken struct {
	UserID    uint   `json:"user_id"`
	Token     string `json:"token"`
	CreatedAt int64  `json:"created_at"`
	IsActive  bool   `json:"is_active"`
}

// RevokeAllUserPermanentTokens 删除用户的所有永久token
func RevokeAllUserPermanentTokens(userID uint) error {
	global.GVA_LOG.Info("开始删除用户所有永久token", zap.Uint("user_id", userID))
	
	// 获取用户的所有永久token
	tokens, err := GetUserPermanentTokens(userID)
	if err != nil {
		global.GVA_LOG.Error("获取用户token列表失败", 
			zap.Uint("user_id", userID),
			zap.Error(err))
		return err
	}
	
	global.GVA_LOG.Info("找到用户token", 
		zap.Uint("user_id", userID),
		zap.Int("total_tokens", len(tokens)))
	
	// 删除所有token
	deletedCount := 0
	ctx := context.Background()
	for _, token := range tokens {
		global.GVA_LOG.Info("准备删除token", 
			zap.Uint("user_id", userID),
			zap.String("token", token.Token))
		
		// 直接从Redis中删除token
		tokenKey := fmt.Sprintf("permanent_token:%s", token.Token)
		err := global.GVA_REDIS.Del(ctx, tokenKey).Err()
		if err != nil {
			global.GVA_LOG.Error("删除用户永久token失败", 
				zap.Uint("user_id", userID),
				zap.String("token", token.Token),
				zap.String("token_key", tokenKey),
				zap.Error(err))
			// 继续删除其他token，不因为单个失败而停止
		} else {
			deletedCount++
			global.GVA_LOG.Info("删除用户永久token成功", 
				zap.Uint("user_id", userID),
				zap.String("token", token.Token),
				zap.String("token_key", tokenKey))
		}
	}
	
	global.GVA_LOG.Info("删除token操作完成", 
		zap.Uint("user_id", userID),
		zap.Int("total_tokens", len(tokens)),
		zap.Int("deleted_tokens", deletedCount))
	
	return nil
}

// GeneratePermanentToken 为用户生成永久token（每个用户只能有一个有效token）
func GeneratePermanentToken(userID uint) (string, error) {
	global.GVA_LOG.Info("开始为用户生成永久token", zap.Uint("user_id", userID))
	
	// 首先获取用户现有的token数量
	existingTokens, err := GetUserPermanentTokens(userID)
	if err != nil {
		global.GVA_LOG.Error("获取用户现有token失败", 
			zap.Uint("user_id", userID),
			zap.Error(err))
	} else {
		global.GVA_LOG.Info("用户当前拥有token数量", 
			zap.Uint("user_id", userID),
			zap.Int("token_count", len(existingTokens)))
	}
	
	// 撤销用户的所有现有token
	err = RevokeAllUserPermanentTokens(userID)
	if err != nil {
		global.GVA_LOG.Error("撤销用户现有永久token失败", 
			zap.Uint("user_id", userID),
			zap.Error(err))
		// 即使撤销失败，也继续生成新token，因为新token会覆盖旧的使用
	} else {
		global.GVA_LOG.Info("成功撤销用户所有现有token", zap.Uint("user_id", userID))
	}
	
	// 生成随机字节
	randomBytes := make([]byte, 32)
	_, err = rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	// 创建token字符串：用户ID + 时间戳 + 随机字符串
	timestamp := time.Now().Unix()
	tokenData := fmt.Sprintf("%d_%d_%s", userID, timestamp, hex.EncodeToString(randomBytes))
	
	// 使用MD5生成最终token
	hash := md5.Sum([]byte(tokenData))
	token := hex.EncodeToString(hash[:])
	
	// 添加前缀以区分永久token
	finalToken := fmt.Sprintf("perm_%s", token)
	
	// 存储到Redis中，永久有效（设置一个很长的过期时间，比如10年）
	tokenKey := fmt.Sprintf("permanent_token:%s", finalToken)
	
	// 将token信息存储到Redis，设置10年过期时间
	ctx := context.Background()
	err = global.GVA_REDIS.HSet(ctx, tokenKey, map[string]interface{}{
		"user_id":    userID,
		"token":      finalToken,
		"created_at": timestamp,
		"is_active":  true,
	}).Err()
	
	if err != nil {
		global.GVA_LOG.Error("存储永久token失败", zap.Error(err))
		return "", err
	}
	
	// 设置过期时间为10年
	err = global.GVA_REDIS.Expire(ctx, tokenKey, time.Hour*24*365*10).Err()
	if err != nil {
		global.GVA_LOG.Error("设置永久token过期时间失败", zap.Error(err))
		return "", err
	}
	
	global.GVA_LOG.Info("生成永久token成功（已撤销旧token）", 
		zap.Uint("user_id", userID), 
		zap.String("token", finalToken))
	
	// 验证：确保用户现在只有一个激活的token
	finalTokens, err := GetUserPermanentTokens(userID)
	if err != nil {
		global.GVA_LOG.Error("验证token单一性时获取用户token失败", 
			zap.Uint("user_id", userID),
			zap.Error(err))
	} else {
		// 所有存在的token都是有效的（失效的已被删除）
		tokenCount := len(finalTokens)
		global.GVA_LOG.Info("token生成后验证", 
			zap.Uint("user_id", userID),
			zap.Int("total_token_count", tokenCount),
			zap.String("new_token", finalToken))
		
		if tokenCount > 1 {
			global.GVA_LOG.Warn("警告：用户拥有多个token", 
				zap.Uint("user_id", userID),
				zap.Int("token_count", tokenCount))
		}
	}
	
	return finalToken, nil
}

// ValidatePermanentToken 验证永久token
func ValidatePermanentToken(token string) (*PermanentToken, error) {
	// 检查token格式
	if !strings.HasPrefix(token, "perm_") {
		return nil, fmt.Errorf("无效的永久token格式")
	}
	
	// 从Redis获取token信息
	tokenKey := fmt.Sprintf("permanent_token:%s", token)
	ctx := context.Background()
	tokenData, err := global.GVA_REDIS.HGetAll(ctx, tokenKey).Result()
	
	if err != nil {
		global.GVA_LOG.Error("获取永久token信息失败", zap.Error(err))
		return nil, fmt.Errorf("token验证失败")
	}
	
	if len(tokenData) == 0 {
		return nil, fmt.Errorf("token不存在或已过期")
	}
	
	// 所有存在的token都是有效的（失效的token已被删除）
	
	// 解析用户ID
	var userID uint
	if _, err := fmt.Sscanf(tokenData["user_id"], "%d", &userID); err != nil {
		return nil, fmt.Errorf("无效的用户ID")
	}
	
	// 解析创建时间
	var createdAt int64
	if _, err := fmt.Sscanf(tokenData["created_at"], "%d", &createdAt); err != nil {
		return nil, fmt.Errorf("无效的创建时间")
	}
	
	permanentToken := &PermanentToken{
		UserID:    userID,
		Token:     token,
		CreatedAt: createdAt,
		IsActive:  true,
	}
	
	return permanentToken, nil
}

// RevokePermanentToken 删除永久token
func RevokePermanentToken(token string) error {
	tokenKey := fmt.Sprintf("permanent_token:%s", token)
	
	// 从Redis中完全删除token
	ctx := context.Background()
	err := global.GVA_REDIS.Del(ctx, tokenKey).Err()
	if err != nil {
		global.GVA_LOG.Error("删除永久token失败", 
			zap.String("token", token),
			zap.String("token_key", tokenKey),
			zap.Error(err))
		return err
	}
	
	global.GVA_LOG.Info("删除永久token成功", 
		zap.String("token", token),
		zap.String("token_key", tokenKey))
	return nil
}

// GetUserPermanentTokens 获取用户的所有永久token
func GetUserPermanentTokens(userID uint) ([]PermanentToken, error) {
	// 通过模式匹配查找该用户的所有永久token
	pattern := "permanent_token:perm_*"
	ctx := context.Background()
	keys, err := global.GVA_REDIS.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}
	
	var tokens []PermanentToken
	for _, key := range keys {
		tokenData, err := global.GVA_REDIS.HGetAll(ctx, key).Result()
		if err != nil {
			continue
		}
		
		var tokenUserID uint
		if _, err := fmt.Sscanf(tokenData["user_id"], "%d", &tokenUserID); err != nil {
			continue
		}
		
		// 只返回该用户的token
		if tokenUserID == userID {
			var createdAt int64
			fmt.Sscanf(tokenData["created_at"], "%d", &createdAt)
			
			// 所有存在的token都是有效的（已删除失效token）
			tokens = append(tokens, PermanentToken{
				UserID:    tokenUserID,
				Token:     tokenData["token"],
				CreatedAt: createdAt,
				IsActive:  true, // 所有存在的token都是有效的
			})
		}
	}
	
	return tokens, nil
}