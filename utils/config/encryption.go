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
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
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
	shouldEncrypt, err := shouldEncrypt(config)
	if err != nil || !shouldEncrypt {
		return err
	}
	config.EncryptionIndex = strconv.Itoa(latestEncryptionIndex)
	return handleSecrets(config, encrypt, latestEncryptionIndex)
}

func (config *ConfigV1) decrypt() error {
	// Return if encryption was not initialized, or if config is not encrypted.
	if !isEncryptionInitialized() || config.EncryptionIndex == "" {
		return nil
	}
	encIndex, err := strconv.Atoi(config.EncryptionIndex)
	if err != nil {
		return errorutils.CheckError(err)
	}
	return handleSecrets(config, decrypt, encIndex)
}

func shouldEncrypt(config *ConfigV1) (bool, error) {
	// Cannot encrypt if encryption was not initialized.
	if !isEncryptionInitialized() {
		return false, nil
	}

	disableEncryption, err := utils.GetBoolEnvValue(cliutils.DisableEncryption, false)
	// Encryption is disabled if instructed by env var and the config file was not encrypted yet.
	return !(disableEncryption && config.EncryptionIndex == ""), errorutils.CheckError(err)
}

// Encrypts the config file if needed (if not encrypted, or if encrypted with old key) and saves it.
func updateConfigFileEncryption(config *ConfigV1) error {
	shouldEncrypt, err := shouldEncrypt(config)
	if err != nil || !shouldEncrypt {
		return err
	}

	// If config file is encrypted with the latest index, no action required.
	if config.EncryptionIndex != "" {
		encIndex, err := strconv.Atoi(config.EncryptionIndex)
		if err != nil {
			return errorutils.CheckError(err)
		}
		if encIndex == latestEncryptionIndex {
			return nil
		}
	}

	// Encrypting the config file.
	// Marshalling and unmarshalling to get a new separate config struct, to prevent modifying the rest of the command.
	decryptedContent, err := config.getContent()
	if err != nil {
		return err
	}
	tmpEncConfig := new(ConfigV1)
	err = json.Unmarshal(decryptedContent, &tmpEncConfig)
	if err != nil {
		return errorutils.CheckError(err)
	}
	return saveConfig(tmpEncConfig)
}

func isEncryptionInitialized() bool {
	return EncryptionFile != ""
}

// Encrypt/Decrypt all secrets in the provided config, with the encryption key in the requested index.
func handleSecrets(config *ConfigV1, handler secretHandler, encryptionIndex int) error {
	err := initEncryptionKeys()
	if encryptionIndex > len(encryptionKeys)-1 {
		return errorutils.CheckError(errors.New("encryption index out of range"))
	}
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

// Reading encryption keys from the provided file, if not read already.
func initEncryptionKeys() error {
	// Already Unmarshalled.
	if len(encryptionKeys) > 0 {
		return nil
	}

	encryption := new(Encryption)
	content := []byte(EncryptionFile)
	err := json.Unmarshal(content, &encryption)
	if err != nil {
		return errorutils.CheckError(err)
	}
	encryptionKeys = encryption.Keys
	return nil
}

func encrypt(secret string, encryptionIndex int) (string, error) {
	if secret == "" {
		return "", nil
	}

	key := []byte(*encryptionKeys[encryptionIndex])
	c, err := aes.NewCipher(key)
	if err != nil {
		return "", errorutils.CheckError(err)
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return "", errorutils.CheckError(err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", errorutils.CheckError(err)
	}

	sealed := gcm.Seal(nonce, nonce, []byte(secret), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

func decrypt(encryptedSecret string, encryptionIndex int) (string, error) {
	key := []byte(*encryptionKeys[encryptionIndex])
	cipherText, err := base64.StdEncoding.DecodeString(encryptedSecret)
	if err != nil {
		return "", errorutils.CheckError(err)
	}

	c, err := aes.NewCipher(key)
	if err != nil {
		return "", errorutils.CheckError(err)
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return "", errorutils.CheckError(err)
	}

	nonceSize := gcm.NonceSize()
	if len(cipherText) < nonceSize {
		return "", errorutils.CheckError(errors.New("unexpected cipher text size"))
	}

	nonce, cipherText := cipherText[:nonceSize], cipherText[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", errorutils.CheckError(err)
	}
	return string(plaintext), nil
}
