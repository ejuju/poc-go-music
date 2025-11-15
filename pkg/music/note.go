package music

import (
	"math"
	"time"
)

// Transposes a frequency up or down a given number of semitones (according to the equal tempered scale).
func Transpose(freq float64, semitones float64) float64 {
	var c = math.Pow(2, 1.0/12.0)
	return float64(freq) * math.Pow(c, semitones)
}

type Note int

func (n Note) Hz() float64                    { return Transpose(440, float64(n)) }
func (n Note) At(x time.Duration) (y float64) { return n.Hz() }

const (
	A4 = Note(iota)
	Bb4
	B4
	C4
	Db4
	D4
	Eb4
	E4
	F4
	Gb4
	G4
	Ab4
)
