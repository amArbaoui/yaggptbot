package storage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
)

type EncryptionService struct {
	encryptionKey []byte
}

func NewEncryptionService(encryptionKey string) *EncryptionService {
	decoded, err := DecodeKeyFromString(encryptionKey)
	if err != nil {
		panic("failed to set up encryption key")
	}
	return &EncryptionService{encryptionKey: decoded}

}

func (e *EncryptionService) Encrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.encryptionKey)
	if err != nil {
		log.Printf("failed to encrypt data, %s", err)
		return nil, err
	}
	cipherData := make([]byte, aes.BlockSize+len(data))

	iv := cipherData[:aes.BlockSize]

	// See https://gist.github.com/josephspurrier/12cc5ed76d2228a41ceb
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		log.Printf("failed to encrypt data, %s", err)
		return nil, err
	}
	stream := cipher.NewCFBEncrypter(block, iv)

	stream.XORKeyStream(cipherData[aes.BlockSize:], data)

	return cipherData, nil
}

func (e *EncryptionService) Decrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.encryptionKey)
	if err != nil {
		return nil, err
	}
	if len(data) < aes.BlockSize {
		err = fmt.Errorf("text is too short")
		log.Printf("failed to decrypt, %s", err)
		return nil, err
	}

	iv := data[:aes.BlockSize]

	data = data[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)

	stream.XORKeyStream(data, data)

	return data, nil
}

func GenerateAESKey(byteSize int) ([]byte, error) {
	key := make([]byte, byteSize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	return key, nil
}

func EncodeKeyToString(key []byte) string {
	return base64.StdEncoding.EncodeToString(key)
}

func DecodeKeyFromString(keyStr string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(keyStr)
}
