package momentType

import "time"

type UserMomentItem struct {
	UserRequest
	RegisterTime time.Time
}

type UserRequest struct {
	UserId          string `json:"id"`
	RedisMomentBody `json:"momentData"`
}

type RedisMomentBody struct {
	Accuracy     int       `json:"accuracy" redis:"accuracy"`
	Activity     string    `json:"activity" redis:"activity"`
	AreaLandedAt string    `json:"areaLandedAt" redis:"areaLandedAt"`
	Battery      int       `json:"battery" redis:"battery"`
	Heading      float64   `json:"heading" redis:"heading"`
	IsCharging   bool      `json:"isCharging" redis:"isCharging"`
	IsMoving     bool      `json:"isMoving" redis:"isMoving"`
	Status       string    `json:"status" redis:"status"`
	LatLng       []float64 `json:"latLng" redis:"latLng"`
	MovingSpeed  int       `json:"movingSpeed" redis:"movingSpeed"`
}

var MomentQueue = make([]UserMomentItem, 0, 100000)

var RegisterInterval = 10 * time.Second
