package storm

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type Inventory struct{}

func (i *Inventory) Load(file string) (*InventoryConfig, error) {
	fileContent, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	// Unmarshal the file content into the InventoryConfig struct
	config := &InventoryConfig{}
	err = yaml.Unmarshal(fileContent, &config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (i *Inventory) aes(encryptionKey string) (cipher.AEAD, error) {
	key := []byte(encryptionKey)
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, errors.New("encryption key must be 16, 24, or 32 bytes long")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return gcm, nil
}

func (i *Inventory) Encrypt(file string, encryptionKey string) (*string, error) {
	fileContent, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	gcm, err := i.aes(encryptionKey)
	if err != nil {
		return nil, err
	}

	// Create a nonce of the correct size
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Encrypt and prepend nonce
	ciphertext := gcm.Seal(nonce, nonce, fileContent, nil)

	// Encode the ciphertext to Base64 for safe storage
	encodedCiphertext := base64.StdEncoding.EncodeToString(ciphertext)
	return &encodedCiphertext, nil
}

func (i *Inventory) Decrypt(ciphertext string, decryptionKey string) (*string, error) {
	gcm, err := i.aes(decryptionKey)
	if err != nil {
		return nil, err
	}

	// Decode the Base64 ciphertext
	ciphertextBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertextBytes) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	// Separate nonce and actual ciphertext
	nonce, encryptedMessage := ciphertextBytes[:nonceSize], ciphertextBytes[nonceSize:]

	// Decrypt the ciphertext
	plaintext, err := gcm.Open(nil, nonce, encryptedMessage, nil)
	if err != nil {
		return nil, err
	}

	decryptedFileContent := string(plaintext)
	return &decryptedFileContent, nil
}

func NewInventory() *Inventory {
	return &Inventory{}
}
