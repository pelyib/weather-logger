package mq

import "github.com/pelyib/weather-logger/internal/shared"

type MsgBody struct {
	Loc shared.Location `json:"loc"`
}
