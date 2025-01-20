package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func init() {
	// Look for .env file in the client directory
	clientDir := filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(os.Args[0]))), "client")
	err := godotenv.Load(filepath.Join(clientDir, ".env"))
	if err != nil {
		// If not found in client dir, try current directory
		err = godotenv.Load()
		if err != nil {
			// If still not found, try parent directory
			err = godotenv.Load(filepath.Join("..", "..", ".env"))
			if err != nil {
				panic("Error loading .env file")
			}
		}
	}
}

// EncryptMessage encrypts a message using AES-256 and returns base64 encoded string
func EncryptMessage(message string) (string, error) {
	secretKey := []byte(os.Getenv("SECRET_KEY"))
	if len(secretKey) == 0 {
		return "", errors.New("SECRET_KEY not found in environment")
	}

	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return "", err
	}

	plaintext := []byte(message)
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

// DecryptMessage decrypts a base64 encoded encrypted message
func DecryptMessage(encrypted string) (string, error) {
	secretKey := []byte(os.Getenv("SECRET_KEY"))
	if len(secretKey) == 0 {
		return "", errors.New("SECRET_KEY not found in environment")
	}

	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return "", err
	}

	ciphertext, err := base64.URLEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	if len(ciphertext) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return string(ciphertext), nil
}
