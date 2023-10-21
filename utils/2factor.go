package utils

import (
	"bytes"
	"encoding/base32"
	"fmt"
	"image/png"
	"os"

	"github.com/pquerna/otp/totp"
	"gorm.io/gorm"
)

// The 2factor logic belows works with the Google authenticator mobile app.
func ShowQrCode(email string, db *gorm.DB) ([]byte, error) {
	var data []byte
	var CIPER_SECRET string = os.Getenv("ENCRYPTION_KEY")

	secret := base32.StdEncoding.EncodeToString([]byte(CIPER_SECRET))
	passcode, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "PlexScriptables",
		AccountName: email,
		Secret:      []byte(secret),
	})

	if err != nil {
		return data, err

	}

	var buf bytes.Buffer
	img, err := passcode.Image(200, 200)
	if err != nil {
		fmt.Println(err)
	}
	png.Encode(&buf, img)

	secret = Encrypt(passcode.Secret())

	db.Table("users").Where("email=?", email).Update("two_factor_code", secret)

	return buf.Bytes(), nil
}
