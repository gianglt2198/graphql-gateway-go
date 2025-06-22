package helpers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"

	"go.uber.org/fx"

	"github.com/gianglt2198/federation-go/package/config"
)

type (
	AESCipher struct {
		gcm cipher.AEAD
	}

	Encryptor interface {
		Encrypt(data []byte) ([]byte, error)
		Decrypt(data []byte) ([]byte, error)
	}
)

type EncryptorHelperParams struct {
	fx.In

	Config config.EncryptConfig
}

func NewAESCipher(params EncryptorHelperParams) Encryptor {
	var gcm cipher.AEAD

	block, err := aes.NewCipher([]byte(params.Config.SecretKey))
	if err != nil {
		panic("failed to create new cipher: " + err.Error())
	}

	gcm, err = cipher.NewGCM(block)
	if err != nil {
		panic("failed to create new GCM: " + err.Error())
	}

	return &AESCipher{gcm: gcm}
}

func (c *AESCipher) Encrypt(data []byte) ([]byte, error) {
	// We need a 12-byte nonce for GCM (modifiable if you use cipher.NewGCMWithNonceSize())
	// A nonce should always be randomly generated for every encryption.
	nonce := make([]byte, c.gcm.NonceSize())
	_, err := rand.Read(nonce)
	if err != nil {
		return nil, err
	}

	// ciphertext here is actually nonce+ciphertext
	// So that when we decrypt, just knowing the nonce size
	// is enough to separate it from the ciphertext.
	encrypted := c.gcm.Seal(nonce, nonce, data, nil)

	return encrypted, nil
}

func (c *AESCipher) Decrypt(data []byte) ([]byte, error) {
	// Since we know the ciphertext is actually nonce+ciphertext
	// And len(nonce) == NonceSize(). We can separate the two.
	nonceSize := c.gcm.NonceSize()
	nonce, data := data[:nonceSize], data[nonceSize:]

	decrypted, err := c.gcm.Open(nil, []byte(nonce), data, nil)
	if err != nil {
		return nil, err
	}

	return decrypted, nil
}
