package strgen

import (
	"github.com/elgs/gostrgen"
)

//RandStr 生成随机字符串
func RandStr(len int) string {
	charSet := gostrgen.LowerUpper | gostrgen.Digit
	includes := ""   // optionally include some additional letters
	excludes := "Ol" //exclude big 'O' and small 'l' to avoid confusion with zero and one.

	if str, err := gostrgen.RandGen(len, charSet, includes, excludes); err == nil {
		return str
	} else {
		return ""
	}
}
