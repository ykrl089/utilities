/*
* @Author: guodi
* @Date:   2021-02-19 23:03:07
* @Last Modified by:   guodi
* @Last Modified time: 2021-02-19 23:04:09
 */
package snowflake

import (
	"github.com/sony/sonyflake"
	"fmt"
	"time"
)

var snowflakeID *sonyflake.Sonyflake

func init() {
	t, _ := time.Parse("2006-01-02", "2017-01-01")
	snowflakeSet := sonyflake.Settings{StartTime: t, MachineID: MachineId}
	snowflakeID = sonyflake.NewSonyflake(snowflakeSet)
	if snowflakeID == nil {
		panic("snow flake ID创建错误！")
	}
}
func SnowflakeId() string {
	id, er := snowflakeID.NextID()
	if er != nil {
		fmt.Println("error ", er.Error())
	}
	return fmt.Sprintf("%d", id)
}
