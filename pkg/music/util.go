package music

import "time"

type BPM float64

func (v BPM) T(beats float64) time.Duration {
	return time.Duration(beats * float64(time.Minute) / float64(v))
}
