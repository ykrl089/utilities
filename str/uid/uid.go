package uid

import (
	"utilities/str/uid/strgen"
	"utilities/str/uid/uuid"
)

func Rand(len int) string {
	return strgen.RandStr(len)
}

func UUID() string {
	return uuid.UUID()
}
