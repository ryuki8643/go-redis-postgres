package gin_server

import (
	"github.com/geolocket/batch_redis/internal/momentType"
	"time"
)

func UserMomentWriteInQueue(request momentType.UserRequest) {
	m := momentType.UserMomentItem{UserRequest: request, RegisterTime: time.Now()}
	momentType.MomentQueue = append(momentType.MomentQueue, m)
}
