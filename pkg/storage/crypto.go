package storage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"os"
)

func GetDefaultRandomKey() ([]byte, error) {
	return generateRandomKey(DefaultRandomKeySize)
}

func generateRandomKey(keySize int) ([]byte, error) {
	key := make([]byte, keySize)
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func encrypt(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, aes.BlockSize+len(data))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], data)

	return ciphertext, nil
}

func decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	// 使用AES加密算法中的CTR模式
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return ciphertext, nil
}

func EncryptFile(aeskey, rawPath, dstPath string) error {
	fileContent, err := os.ReadFile(rawPath)
	if err != nil {
		return err
	}
	fileInfo, err := os.Stat(rawPath)
	if err != nil {
		return err
	}
	aeskeyBytes := []byte(aeskey)
	encryptedContent, err := encrypt(fileContent, aeskeyBytes)
	if err != nil {
		return err
	}
	err = os.WriteFile(dstPath, encryptedContent, fileInfo.Mode())
	if err != nil {
		return err

	}
	return nil
}

func DecryptFile(aeskey, encryptedPath, dstPath string) error {
	encryptedContent, err := os.ReadFile(encryptedPath)
	if err != nil {
		return err
	}
	fileInfo, err := os.Stat(encryptedPath)
	if err != nil {
		return err
	}
	aeskeyBytes, err := hex.DecodeString(aeskey)
	if err != nil {
		return err
	}
	decryptedContent, err := decrypt(encryptedContent, aeskeyBytes)
	if err != nil {
		return err
	}
	err = os.WriteFile(dstPath, decryptedContent, fileInfo.Mode())
	if err != nil {
		return err
	}
	return nil
}
