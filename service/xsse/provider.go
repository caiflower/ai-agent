package xsse

import (
	"time"

	"github.com/tmaxmax/go-sse"
)

func NewSSEProvider() sse.Provider {
	rp, _ := sse.NewValidReplayer(time.Minute*2, true)
	rp.GCInterval = time.Minute
	return &sse.Joe{Replayer: rp}
}
