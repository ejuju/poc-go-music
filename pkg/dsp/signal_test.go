package dsp

import (
	"fmt"
	"math"
	"testing"
	"time"
)

const maxDiff = 0.00000001

func TestSignal(t *testing.T) {
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
		{
			s: Sine(Constant(2.0)),
			want: map[time.Duration]float64{
				0:                      0,
				125 * time.Millisecond: 1,
				250 * time.Millisecond: 0,
				375 * time.Millisecond: -1,
				500 * time.Millisecond: 0,
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

func TestOscillator(t *testing.T) {
	s := Sine(Constant(1000))
	periodDuration := time.Second / 1000
	for x := time.Duration(0); x < time.Second; x += time.Millisecond {
		a := s.At(x)
		b := s.At(x + periodDuration)
		if math.Abs(a-b) > maxDiff {
			t.Fatalf("s(%s)=%f but s(%s+%s)=%f", x, a, x, periodDuration, b)
		}
	}
}
