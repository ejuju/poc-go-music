package dsp

import (
	"fmt"
	"math"
	"testing"
	"time"
)

func TestSignal(t *testing.T) {
	const maxDiff = 0.00000001

	tests := []struct {
		s    Signal
		want map[time.Duration]float64
	}{
		{
			s: Constant(420),
			want: map[time.Duration]float64{
				0:                     420,
				69 * time.Millisecond: 420,
				time.Hour:             420,
			},
		},
		{
			s: Sine(Constant(1.0)),
			want: map[time.Duration]float64{
				0:                      0,
				250 * time.Millisecond: 1,
				500 * time.Millisecond: 0,
				750 * time.Millisecond: -1,
				time.Second:            0,
			},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			for x, y := range test.want {
				got := test.s.At(x)
				if math.Abs(got-y) > maxDiff {
					t.Fatalf("got %f not %f (at %s)", got, y, x)
				}
			}
		})
	}
}
