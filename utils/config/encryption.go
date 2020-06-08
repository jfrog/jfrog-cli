package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"os"
	"strconv"
	"syscall"
)

type SecurityConf struct {
	Version   string `yaml:"version,omitempty"`
	MasterKey string `yaml:"masterKey,omitempty"`
}

const masterKeyField = "masterKey"
const masterKeyLength = 32
const encryptErrorPrefix = "cannot encrypt config: "
const decryptErrorPrefix = "cannot decrypt config: "

type secretHandler func(string, string) (string, error)

// Encrypt config file if security configuration file exists and contains master key.
func (config *ConfigV2) encrypt() error {
	key, _, err := getMasterKeyFromSecurityConfFile()
	if err != nil || key == "" {
		return err
	}
	// Mark config as encrypted.
	config.Enc = true
	return handleSecrets(config, encrypt, key)
}

// Decrypt config if encrypted and master key exists.
func (config *ConfigV2) decrypt() error {
	if !config.Enc {
		return updateEncryptionIfNeeded(config)
	}
	key, secFileExists, err := getMasterKeyFromSecurityConfFile()
	if err != nil {
		return err
	}
	if !secFileExists {
		return errorutils.CheckError(errors.New(decryptErrorPrefix + "security configuration file was not found"))
	}
	if key == "" {
		return errorutils.CheckError(errors.New(decryptErrorPrefix + "security configuration file does not contain a master key"))
	}
	return handleSecrets(config, decrypt, key)
}

// Encrypt the config file if it is decrypted while security configuration file exists and contains a master key.
func updateEncryptionIfNeeded(originalConfig *ConfigV2) error {
	masterKey, _, err := getMasterKeyFromSecurityConfFile()
	if err != nil || masterKey == "" {
		return err
	}

	// Marshalling and unmarshalling to get a new separate config struct, to prevent modifying the config for the rest of the execution.
	decryptedContent, err := originalConfig.getContent()
	if err != nil {
		return err
	}
	tmpEncConfig := new(ConfigV2)
	err = json.Unmarshal(decryptedContent, &tmpEncConfig)
	if err != nil {
		return errorutils.CheckError(err)
	}
	err = saveConfig(tmpEncConfig)
	if err != nil {
		return err
	}
	// Mark that config file is encrypted
	originalConfig.Enc = true
	return nil
}

// Encrypt/Decrypt all secrets in the provided config, with the provided master key.
func handleSecrets(config *ConfigV2, handler secretHandler, key string) error {
	var err error
	for _, rtDetails := range config.Artifactory {
		rtDetails.Password, err = handler(rtDetails.Password, key)
		if err != nil {
			return err
		}
		rtDetails.AccessToken, err = handler(rtDetails.AccessToken, key)
		if err != nil {
			return err
		}
		rtDetails.ApiKey, err = handler(rtDetails.ApiKey, key)
		if err != nil {
			return err
		}
		rtDetails.SshPassphrase, err = handler(rtDetails.SshPassphrase, key)
		if err != nil {
			return err
		}
		rtDetails.RefreshToken, err = handler(rtDetails.RefreshToken, key)
		if err != nil {
			return err
		}
	}
	if config.Bintray != nil {
		config.Bintray.Key, err = handler(config.Bintray.Key, key)
		if err != nil {
			return err
		}
	}
	if config.MissionControl != nil {
		config.MissionControl.AccessToken, err = handler(config.MissionControl.AccessToken, key)
		if err != nil {
			return err
		}
	}
	return nil
}

func getMasterKeyFromSecurityConfFile() (key string, secFileExists bool, err error) {
	secFile, err := cliutils.GetJfrogSecurityConfFilePath()
	if err != nil {
		return "", false, err
	}
	exists, err := fileutils.IsFileExists(secFile, false)
	if err != nil || !exists {
		return "", false, err
	}

	config := viper.New()
	config.SetConfigType("yaml")
	f, err := os.Open(secFile)
	if err != nil {
		return "", false, errorutils.CheckError(err)
	}
	err = config.ReadConfig(f)
	if err != nil {
		return "", false, errorutils.CheckError(err)
	}
	key = config.GetString(masterKeyField)
	return key, true, nil
}

func readMasterKeyFromConsole() (string, error) {
	print("Please enter the master key: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", errorutils.CheckError(err)
	}
	// New-line required after the input:
	fmt.Println()
	return string(bytePassword), nil
}

func encrypt(secret string, key string) (string, error) {
	if secret == "" {
		return "", nil
	}
	if len(key) != 32 {
		return "", errorutils.CheckError(errors.New(encryptErrorPrefix + "Wrong length for master key. Key should have a length of exactly: " + strconv.Itoa(masterKeyLength) + " bytes"))
	}
	c, err := aes.NewCipher([]byte(key))
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

func decrypt(encryptedSecret string, key string) (string, error) {
	if encryptedSecret == "" {
		return "", nil
	}

	cipherText, err := base64.StdEncoding.DecodeString(encryptedSecret)
	if err != nil {
		return "", errorutils.CheckError(err)
	}

	c, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", errorutils.CheckError(err)
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return "", errorutils.CheckError(err)
	}

	nonceSize := gcm.NonceSize()
	if len(cipherText) < nonceSize {
		return "", errorutils.CheckError(errors.New(decryptErrorPrefix + "unexpected cipher text size"))
	}

	nonce, cipherText := cipherText[:nonceSize], cipherText[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", errorutils.CheckError(err)
	}
	return string(plaintext), nil
}
