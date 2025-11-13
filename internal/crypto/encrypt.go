package crypto

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "fmt"
    "io"

    "golang.org/x/crypto/scrypt"
)

// DeriveKey - scrypt(Passphrase, Salt) → 32-byte key
func DeriveKey(passphrase string, salt []byte) ([]byte, error) {
	return scrypt.Key([]byte(passphrase), salt, 32768, 8, 1, 32)
}

// Encrypt - AES-256-GCM
func Encrypt(plainText []byte, passphrase string) (cipherText, iv []byte, err error) {
	salt := make([]byte, 8)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, nil, err
	}

	key, err := DeriveKey(passphrase, salt)
	if err != nil {
		return nil, nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	iv = make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, nil, err
	}

	// Prepend salt
	fullCipher := gcm.Seal(nil, iv, plainText, nil)
	cipherText = append(salt, iv...)
	cipherText = append(cipherText, fullCipher...)

	return cipherText, iv, nil
}

// Decrypt - reverse of Encrypt
func Decrypt(cipherText []byte, passphrase string) ([]byte, error) {
	if len(cipherText) < 8+12 {
		return nil, fmt.Errorf("ciphertext too short")
	}

	salt := cipherText[:8]
	iv := cipherText[8:20]
	data := cipherText[20:]

	key, err := DeriveKey(passphrase, salt)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plainText, err := gcm.Open(nil, iv, data, nil)
	if err != nil {
		return nil, err
	}

	return plainText, nil
}
