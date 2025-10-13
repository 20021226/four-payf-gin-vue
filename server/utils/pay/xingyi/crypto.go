package xingyi

import (
    "crypto/aes"
    "errors"
    "encoding/base64"
)

func pkcs7Pad(data []byte, blockSize int) []byte {
    padLen := blockSize - (len(data) % blockSize)
    if padLen == 0 { padLen = blockSize }
    pad := make([]byte, padLen)
    for i := range pad { pad[i] = byte(padLen) }
    return append(data, pad...)
}

// ecbEncrypt performs AES-ECB encryption
func ecbEncrypt(key, data []byte) ([]byte, error) {
    if len(key) != 16 && len(key) != 24 && len(key) != 32 {
        return nil, errors.New("invalid AES key length")
    }
    block, err := aes.NewCipher(key)
    if err != nil { return nil, err }
    bs := block.BlockSize()
    if len(data)%bs != 0 { return nil, errors.New("data not full blocks") }
    out := make([]byte, len(data))
    for i := 0; i < len(data); i += bs {
        block.Encrypt(out[i:i+bs], data[i:i+bs])
    }
    return out, nil
}

// EncryptAES_ECB_PKCS7 matches Python AES(ECB, PKCS7) then base64
func EncryptAES_ECB_PKCS7(word string) (string, error) {
    key := []byte("abcdefgabcdefg12")
    plain := []byte(word)
    padded := pkcs7Pad(plain, aes.BlockSize)
    enc, err := ecbEncrypt(key, padded)
    if err != nil { return "", err }
    return base64.StdEncoding.EncodeToString(enc), nil
}