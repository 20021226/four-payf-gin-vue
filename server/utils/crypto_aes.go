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
