package storage

import (
	"testing"
)

var testData = "test-data"

func EncryptionServiceTest(t *testing.T) {
	secretKey, err := GenerateAESKey(32)
	if err != nil {
		t.Errorf("key should be generated")
		t.FailNow()
	}
	encodedKey := EncodeKeyToString(secretKey)
	encryptionService := NewEncryptionService(encodedKey)
	encrypted, err := encryptionService.Encrypt([]byte(testData))
	if err != nil {
		t.Errorf("encrypt should work %s", err)
		t.FailNow()
	}
	decrypted, err := encryptionService.Decrypt(encrypted)
	if err != nil {
		t.Errorf("decrypt should work %s", err)
		t.FailNow()
	}
	decryptedStr := string(decrypted)
	if decryptedStr != testData {
		t.Errorf("test data mutated after encrypt and decrypt operations")
	}
}
