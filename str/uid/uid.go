package uid

import (
	"github.com/ykrl089/utilities/str/uid/snowflake"
	"github.com/ykrl089/utilities/str/uid/strgen"
	"github.com/ykrl089/utilities/str/uid/uuid"
)

func Rand(len int) string {
	return strgen.RandStr(len)
}

func UUID() string {
	return uuid.UUID()
}

func SnowflakeID() string {
	return snowflake.SnowflakeId()
}
