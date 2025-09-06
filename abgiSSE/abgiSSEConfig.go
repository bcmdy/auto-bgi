package abgiSSE

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
)

// 解密
func Decrypt(encryptedText, key string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return "", errors.New("key length must be 16, 24, or 32 bytes")
	}

	cipherText, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", err
	}

	if len(cipherText) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}

	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherText, cipherText)

	return string(cipherText), nil
}
