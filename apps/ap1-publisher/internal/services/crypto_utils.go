package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

// Clave acordada de 32 bytes (AES-256) compartida con AP2
var SecretKey = []byte("12345678901234567890123456789012")

// EncryptData cifra un slice de bytes en memoria usando AES-GCM
func EncryptData(plainText []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plainText, nil), nil
}

// DecryptData descifra el archivo binario (útil para pruebas)
func DecryptData(cipherText []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(cipherText) < nonceSize {
		return nil, errors.New("datos cifrados inválidos")
	}

	nonce, encryptedBytes := cipherText[:nonceSize], cipherText[nonceSize:]
	return gcm.Open(nil, nonce, encryptedBytes, nil)
}