package utils

import (
	"os"

	"github.com/golang-module/dongle"
)

var cipher *dongle.Cipher

func InitCipher() {
	cipher = dongle.NewCipher()
	cipher.SetMode(dongle.CBC)
	cipher.SetPadding(dongle.PKCS7)
	cipher.SetKey(os.Getenv("ENCRYPT_SECRET_KEY"))
	cipher.SetIV(os.Getenv("ENCRYPT_IV_KEY"))
}

func Encrypt(text string) string {
	return dongle.Encrypt.FromString(text).By3Des(cipher).ToHexString()
}

func Decrypt(text string) string {
	return dongle.Decrypt.FromHexString(text).By3Des(cipher).ToString()
}
