package uuid

import (
	uuid "github.com/satori/go.uuid"
)

// UUID 生成uuid
func UUID() string {
	var err error
	return uuid.Must(uuid.NewV4(), err).String()
}
