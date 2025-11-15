package dsp

import (
	"encoding/binary"
	"math"
	"time"
)

type Signal interface {
	At(x time.Duration) (y float64)
}

type SignalFunc func(x time.Duration) (y float64)

func (f SignalFunc) At(x time.Duration) (y float64) { return f(x) }

func Constant(v float64) Signal {
	return SignalFunc(func(x time.Duration) float64 { return v })
}

func Sine(freq Signal) Signal {
	return SignalFunc(func(x time.Duration) (y float64) {
		return math.Sin(x.Seconds() * 2 * math.Pi * freq.At(x))
	})
}

func Sample(s Signal, rate int, from, to time.Duration) (frames []float64) {
	step := float64(time.Second) / float64(rate)
	for i := float64(from); i < float64(from+to); i += step {
		val := s.At(time.Duration(i))
		frames = append(frames, val)
	}
	return frames
}

func EncodePCM(frames []float64) (b []byte) {
	var buf [8]byte
	for _, pulse := range frames {
		binary.BigEndian.PutUint64(buf[:], math.Float64bits(pulse))
		b = append(b, buf[:]...)
	}
	return b
}

func Combine(signals ...Signal) Signal {
	return SignalFunc(func(x time.Duration) (y float64) {
		for _, s := range signals {
			y += s.At(x)
		}
		return y / float64(len(signals))
	})
}

type FiniteSignal struct {
	Signal
	time.Duration
}

func F(d time.Duration, s Signal) FiniteSignal { return FiniteSignal{s, d} }

func Blank(d time.Duration) FiniteSignal { return FiniteSignal{Constant(0), d} }

func Sequence(signals ...FiniteSignal) Signal {
	totalDuration := time.Duration(0)
	for _, s := range signals {
		totalDuration += s.Duration
	}
	return SignalFunc(func(x time.Duration) (y float64) {
		x = x % totalDuration
		i := time.Duration(0)
		for _, s := range signals {
			if x >= i && x < i+s.Duration {
				return s.Signal.At(x)
			}
			i += s.Duration
		}
		panic("unreachable")
	})
}

func Lerp(from, to float64, over time.Duration) FiniteSignal {
	return F(over, SignalFunc(func(x time.Duration) (y float64) {
		return from + (to-from)*math.Mod(float64(x), float64(over))/float64(over)
	}))
}

func Amplify(v, by Signal) Signal {
	return SignalFunc(func(x time.Duration) (y float64) {
		return v.At(x) * by.At(x)
	})
}
