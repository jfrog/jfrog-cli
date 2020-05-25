package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-client-go/utils"
	"io"
	"strconv"
)
var EncryptionFile string
var encryptionKeys []*string
const latestEncryptionIndex = 0

type Encryption struct {
	Keys    []*string `json:"Keys"`
	Version string    `json:"Version,omitempty"`
}

type secretHandler func(string, int) (string, error)

func (config *ConfigV1) encrypt() error {
	// Not encrypting when explicitly instructed so, or encryption was not initialized.
	disableEncryption, err := utils.GetBoolEnvValue(cliutils.DisableEncryption, false)
	if err != nil || EncryptionFile == "" || disableEncryption {
		config.EncryptionIndex = ""
		return err
	}
	config.EncryptionIndex = strconv.Itoa(latestEncryptionIndex)
	return handleSecrets(config, encryptSecret, latestEncryptionIndex)
}

func (config *ConfigV1) decrypt() error {
	encIndex, err := strconv.Atoi(config.EncryptionIndex)
	if err != nil {
		return err
	}
	return handleSecrets(config, decryptSecret, encIndex)
}

// Decrypts the config struct or encrypts the config file, if needed.
func handleCurrentEncryptionStatus(config *ConfigV1, content []byte) error {
	// Return if encryption was not initialized.
	if EncryptionFile == "" {
		return nil
	}

	disableEncryption, err := utils.GetBoolEnvValue(cliutils.DisableEncryption, false)
	if err != nil {
		return err
	}

	// Already encrypted. Decrypt if necessary.
	if config.EncryptionIndex != "" {
		err = config.decrypt()
		if err != nil {
			return err
		}
		// Save the decrypted config to file if disabling is required.
		if disableEncryption {
			return saveConfig(config)
		}
		return nil
	}

	// Encrypt the config file if necessary.
	if !disableEncryption {
		// Unmarshalling the content again to avoid modifying the config used in the rest of the command.
		tmpEncConfig := new(ConfigV1)
		err = json.Unmarshal(content, &tmpEncConfig)
		if err != nil {
			return err
		}
		return saveConfig(tmpEncConfig)
	}
	return nil
}

func handleSecrets(config *ConfigV1, handler secretHandler, encryptionIndex int) error {
	if encryptionIndex > latestEncryptionIndex {
		return errors.New("invalid encryption index in config")
	}
	err := getEncryptionKeys()
	if err != nil {
		return err
	}
	for _, rtDetails := range config.Artifactory {
		rtDetails.Password, err = handler(rtDetails.Password, encryptionIndex)
		if err != nil {
			return err
		}
		rtDetails.AccessToken, err = handler(rtDetails.AccessToken, encryptionIndex)
		if err != nil {
			return err
		}
		rtDetails.ApiKey, err = handler(rtDetails.ApiKey, encryptionIndex)
		if err != nil {
			return err
		}
		rtDetails.SshPassphrase, err = handler(rtDetails.SshPassphrase, encryptionIndex)
		if err != nil {
			return err
		}
		rtDetails.RefreshToken, err = handler(rtDetails.RefreshToken, encryptionIndex)
		if err != nil {
			return err
		}
	}
	if config.Bintray != nil {
		config.Bintray.Key, err = handler(config.Bintray.Key, encryptionIndex)
		if err != nil {
			return err
		}
	}
	if config.MissionControl != nil {
		config.MissionControl.AccessToken, err = handler(config.MissionControl.AccessToken, encryptionIndex)
		if err != nil {
			return err
		}
	}
	return nil
}

// Getting encryption keys from the encryption file if it wasn't done yet.
func getEncryptionKeys() error {
	// Already Unmarshalled.
	if len(encryptionKeys) > 0 {
		return nil
	}

	encryption := new(Encryption)
	content := []byte(EncryptionFile)
	err := json.Unmarshal(content, &encryption)
	if err != nil {
		return err
	}
	encryptionKeys = encryption.Keys
	return nil
}

func encryptSecret(secret string, encryptionIndex int) (string, error) {
	if secret == "" {
		return "", nil
	}

	key := []byte(*encryptionKeys[encryptionIndex])
	c, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	sealed := gcm.Seal(nonce, nonce, []byte(secret), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

func decryptSecret(encryptedSecret string, encryptionIndex int) (string, error) {
	key := []byte(*encryptionKeys[encryptionIndex])
	cipherText, err := base64.StdEncoding.DecodeString(encryptedSecret)
	if err != nil {
		return "", err
	}

	c, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(cipherText) < nonceSize {
		return "", err
	}

	nonce, cipherText := cipherText[:nonceSize], cipherText[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
