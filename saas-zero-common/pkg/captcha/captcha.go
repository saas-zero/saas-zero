package captcha

import (
	"image/color"
	"time"

	"github.com/mojocn/base64Captcha"
)

type Result struct {
	Id   string
	B64s string
	Code string
}

var (
	defaultDriver = base64Captcha.NewDriverDigit(80, 240, 4, 0.7, 80)
	memStore      = base64Captcha.NewMemoryStore(10240, 10*time.Minute)
)

func Generate() (*Result, error) {
	c := base64Captcha.NewCaptcha(defaultDriver, memStore)
	id, b64s, code, err := c.Generate()
	if err != nil {
		return nil, err
	}
	return &Result{Id: id, B64s: b64s, Code: code}, nil
}

func GenerateCustom(height, width, length int, maxSkew float64, dotCount int, bgColor color.RGBA) (*Result, error) {
	d := base64Captcha.NewDriverDigit(height, width, length, maxSkew, dotCount)
	c := base64Captcha.NewCaptcha(d, memStore)
	id, b64s, code, err := c.Generate()
	if err != nil {
		return nil, err
	}
	return &Result{Id: id, B64s: b64s, Code: code}, nil
}
