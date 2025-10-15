package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
)

// 加密后的数据结构
type EncryptedData struct {
	Data string `json:"data"` // Base64编码的加密数据
	IV   string `json:"iv"`   // Base64编码的IV
}

// AES-GCM 加密（推荐方式）
func EncryptAESGCM(plaintext string, key []byte) (*EncryptedData, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// 生成随机nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// 加密（GCM会自动添加认证标签）
	ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)

	return &EncryptedData{
		Data: base64.StdEncoding.EncodeToString(ciphertext),
		IV:   base64.StdEncoding.EncodeToString(nonce),
	}, nil
}

// AES-GCM 解密
func DecryptAESGCM(encrypted *EncryptedData, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// 解码Base64
	ciphertext, err := base64.StdEncoding.DecodeString(encrypted.Data)
	if err != nil {
		return "", err
	}

	nonce, err := base64.StdEncoding.DecodeString(encrypted.IV)
	if err != nil {
		return "", err
	}

	// 解密
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// GenerateAESKey 生成指定字节长度的 AES 密钥，并以 Base64 字符串返回
// length 推荐使用 16/24/32 对应 AES-128/192/256
func GenerateAESKey(length int) (string, error) {
	key := make([]byte, length)
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	// 使用标准 Base64 编码，便于存储与传输
	return base64.StdEncoding.EncodeToString(key), nil
}
