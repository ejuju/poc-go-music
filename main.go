package main

import (
	"os"

	"github.com/ejuju/poc-go-music/pkg/dsp"
	"github.com/ejuju/poc-go-music/pkg/music"
)

func main() {
	bpm := music.BPM(127)

	chord1 := dsp.Combine(
		dsp.Sine(music.C4),
		dsp.Sine(music.E4),
		dsp.Sine(music.G4),
	)
	chord2 := dsp.Combine(
		dsp.Sine(music.A4),
		dsp.Sine(music.C4),
		dsp.Sine(music.E4),
	)
	chord3 := dsp.Combine(
		dsp.Sine(music.E4),
		dsp.Sine(music.B4),
		dsp.Sine(music.G4),
	)
	chord4 := dsp.Combine(
		dsp.Sine(music.D4),
		dsp.Sine(music.A4),
		dsp.Sine(music.Gb4),
	)

	s := dsp.Sequence(
		dsp.F(bpm.T(4), dsp.Amplify(chord1, dsp.Sequence(dsp.Lerp(0, 1, bpm.T(2)), dsp.Lerp(1, 0, bpm.T(2))))),
		dsp.F(bpm.T(4), dsp.Amplify(chord2, dsp.Sequence(dsp.Lerp(0, 1, bpm.T(2)), dsp.Lerp(1, 0, bpm.T(2))))),
		dsp.F(bpm.T(4), dsp.Amplify(chord3, dsp.Sequence(dsp.Lerp(0, 1, bpm.T(2)), dsp.Lerp(1, 0, bpm.T(2))))),
		dsp.F(bpm.T(4), dsp.Amplify(chord4, dsp.Sequence(dsp.Lerp(0, 1, bpm.T(2)), dsp.Lerp(1, 0, bpm.T(2))))),
	)

	frames := dsp.Sample(s, 44100, 0, bpm.T(16))
	os.Stdout.Write(dsp.EncodePCM(frames))
}
