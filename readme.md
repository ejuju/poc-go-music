# Making music in Go

In order to produce music with our Go code, we need the following:
- An oscillator (like a sine wave that will produce sound at a certain frequency)
- A way to turn the continuous oscillator signal into audio frames
- A way to interpret these audio frames and play them on a speaker
- Utilities to handle musical notes, chords, scales, etc.

Let's begin with our oscillator.
We'll use a simple sine wave for now, which can be defined as follows:

In `pkg/dsp/signal.go`:
```go
func Sine(freq float64, x time.Duration) (y float64) {
    return math.Sin(x.Seconds() * 2 * math.Pi * freq)
}
```

NB: The frequency is in Hertz.

This function returns the value of our signal at the time "x" for the given frequency "freq" (in Hertz).

Now, if we want a speaker to play some sound using our sine function.
We must encode audio frames that our OS will pass on to the speakers.

To extract frames from a continuous signal, we need:
- A source signal
- A sample rate (how often we measure the value of our signal)
- Where to begin and end our sampling within the signal (offset and length)

But... what is a signal?
Simply, a signal is a function that returns a value (between -1 and 1) that fluctuates over time.
Like so:

```go
type Signal func(x time.Duration) (y float64)
```

Now let's refactor our previous sine function so that it returns a Signal function
that we can use later on to encode audio frames.

```go
func Sine(freq float64) Signal {
    return func(x time.Duration) (y float64) {
        return math.Sin(x.Seconds() * 2 * math.Pi * freq)
    }
}
```

Actually, we can even go one step further, let's say we want our oscillator's frequency to also change over time
(like a siren that goes from low to high frequency), then we can also make the input frequency argument to
be a signal:

```go
func Sine(freq Signal) Signal {
    return func(x time.Duration) (y float64) {
        return math.Sin(x.Seconds() * 2 * math.Pi * freq(x))
    }
}
```

But very often we will want our frequency to stay constant, so let's define a helper for that:

```go
func Constant(v float64) Signal {
    return func(x time.Duration) float64 { return v }
}
```

OK, so now we have a generic Signal type that we can use to:
- Generate sound at a given frequency with an oscillator
- Control other signals
- Sample audio frames

The next step is taking our continuous signal (which for now is just a mathemical function)
and getting audio frames:

```go
func Sample(s Signal, rate int, from, to time.Duration) (frames []float64) {
    step := float64(time.Second) / float64(rate)
	for i := float64(from); i < float64(from+to); i += step {
		val := s(time.Duration(i))
		frames = append(frames, val)
	}
	return frames
}
```

Where:
- `s` is our source signal.
- `rate` is the number of frames per second.
- `from` is where to begin measuring our signal.
- `to` is where to stop measuring our signal.

Cool, so now we went from a continuous signal to audio frames (= measurements) of the signal.
This allows computers to handle the input signal and play audio frames on a speaker.

But, how do we play audio frames?

There are several ways to go about this, but today we will be encoding frames
to a file (in a special format called PCM) and we can then use `ffplay` to
play this file on a speaker.

NB: We are not implementing the "playing" part (that's why we use ffplay),
we are only concerned with audio synthesis for now.

So, we said that we could encode our audio frames to PCM to play them with `ffplay`.
But what is PCM?

PCM is a very simple straightforward file format: audio frames are simply appended to the file,
one after another. The file doesn't have any header or anything else.
Since there's no way for the file to provide metadata, we will need to provide some information
to `ffplay`:
- The sample rate (ex: 44100 Hz)
- The frame encoding format (in our case: F64BE, since we encode each frame as a big-endian float64)

Let's write our PCM encoding function:
```go
func EncodePCM(frames []float64) (b []byte) {
    var buf [8]byte
    for _, pulse := range frames {
		binary.BigEndian.PutUint64(buf[:], math.Float64bits(pulse))
		b = append(b, buf[:]...)
	}
	return b
}
```

We're ready to play our first sound!
Let's write a simple `main.go` file to create a 5-second-long audio PCM file that plays a sine oscillator at 440 Hz.

```go
func main() {
    signal := dsp.Sine(dsp.Constant(440))
    frames := dsp.Sample(signal, 44100, 0, 5*time.Second)
    os.Stdout.Write(dsp.EncodePCM(frames))
}
```

Let's run:
```shell
go run . > tmp/test.pcm && ffplay -f f64be -ar 44100 -autoexit -showmode 1 tmp/test.pcm
```

---

Alright, we can play sound, but this is far from music yet.

Let's start with a simple `music` package that will contain musical notes definitions and other utilities
for handling music (such as transposing notes, defining chords, scales, etc).

We want to be able to refer to notes instead of hardcoded frequencies,
so that we can write `music.A4` instead of using `440` Hertz.

In `pkg/music/note.go`
```go
type Note int

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
```

And with these functions we get the corresponding frequency:
```go
// Transposes a frequency up or down a given number of semitones (according to the equal tempered scale).
func Transpose(root float64, semitones float64) float64 {
	var c = math.Pow(2, 1.0/12.0)
	return float64(root) * math.Pow(c, semitones)
}

// Returns the frequency (in Hertz) corresponding to the given note.
func (n Note) Hz() float64 { return Transpose(440, float64(n)) }
```

And one more thing, it would be nice if our `music.Note` could implement `dsp.Signal`.
So that we can use it as a constant signal like this:
```go
dsp.Sine(music.A4)
```

So let's turn our `dsp.Signal` type from a `func` to an interface.
This will allow us to write more concise code when using our library.

In `pkg/dsp/signal.go`
```go
type Signal interface {
	At(x time.Duration) (y float64)
}

type SignalFunc func(x time.Duration) (y float64)

func (f SignalFunc) At(x time.Duration) (y float64) { return f(x) }
```

And we also need to change the code for some of our other DSP functions.
```go
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
```

In `pkg/music/note.go`:
```go
func (n Note) At(x time.Duration) (y float64) { return n.Hz() }
```

Now that we can play single notes, it would be nice if we could play a chord
(= several notes at once).

For example a Major C chord contains the following notes: C, E, and G.

In order to combine several notes playing together in a signal,
we simply need to do the average of the signals, like this:

In `pkg/dsp/signal.go`:
```go
func Combine(signals ...Signal) Signal {
	return SignalFunc(func(x time.Duration) (y float64) {
		for _, s := range signals {
			y += s.At(x)
		}
		return y / float64(len(signals))
	})
}
```

Let's hear our first chord!

In `main.go`:
```go
func main() {
	chord := dsp.Combine(
		dsp.Sine(music.C4),
		dsp.Sine(music.E4),
		dsp.Sine(music.G4),
	)
	frames := dsp.Sample(chord, 44100, 0, 5*time.Second)
	os.Stdout.Write(dsp.EncodePCM(frames))
}
```

OK, we can play a chord, but what about a chord progression.
To do that, we want a way to say: play this signal, and then this one.
But for now, our signals have no finite duration.

Let's change that:

In `pkg/dsp/signal.go`
```go
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
```

So now we can update our main function to play a sequence of 4 chords.

In `main.go`:
```go
func main() {
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
		dsp.F(time.Second, chord1),
		dsp.F(time.Second, chord2),
		dsp.F(time.Second, chord3),
		dsp.F(time.Second, chord4),
	)
	frames := dsp.Sample(s, 44100, 0, 5*time.Second)
	os.Stdout.Write(dsp.EncodePCM(frames))
}
```

However, defining duration in seconds is a bit awkward.
In music, we usually define the tempo of music using a BPM (number of beats per minute).

In `pkg/music/util.go`:
```go
type BPM float64

func (v BPM) T(beats float64) time.Duration {
	return time.Duration(beats * float64(time.Minute) / float64(v))
}
```

In `main.go`, we can now use our BPM to compute durations for how long to play each chord:
```go
func main() {
	bpm := music.BPM(127)
	
	// ...
	
	s := dsp.Sequence(
		dsp.F(bpm.T(4), chord1),
		dsp.F(bpm.T(4), chord2),
		dsp.F(bpm.T(4), chord3),
		dsp.F(bpm.T(4), chord4),
	)

	frames := dsp.Sample(s, 44100, 0, bpm.T(16))
	os.Stdout.Write(dsp.EncodePCM(frames))
}
```

Now, we would like to modulate the amplitude of our chords so that they fade in and out,
as if someone was turning a volume knob up and down.

In order to do that, we start by defining a simple linear interpolation function
(to be able to compute intermediate values from our beginning and ending values).

In `pkg/dsp/lerp.go`:
```go
func Lerp(from, to float64, over time.Duration) Signal {
	return SignalFunc(func(x time.Duration) (y float64) {
		return from + (to-from)*math.Mod(float64(x), float64(over))/float64(over)
	})
}
```

And we also define a helper to modulate the amplitude of a signal:
```go
func Amplify(v, by Signal) Signal {
	return SignalFunc(func(x time.Duration) (y float64) {
		return v.At(x) * by.At(x)
	})
}
```

And adapt our `main.go` to fade each chord in and out:
```go
func main() {
	s := dsp.Sequence(
		dsp.F(bpm.T(4), dsp.Amplify(chord1, dsp.Sequence(dsp.Lerp(0, 1, bpm.T(2)), dsp.Lerp(1, 0, bpm.T(2))))),
		dsp.F(bpm.T(4), dsp.Amplify(chord2, dsp.Sequence(dsp.Lerp(0, 1, bpm.T(2)), dsp.Lerp(1, 0, bpm.T(2))))),
		dsp.F(bpm.T(4), dsp.Amplify(chord3, dsp.Sequence(dsp.Lerp(0, 1, bpm.T(2)), dsp.Lerp(1, 0, bpm.T(2))))),
		dsp.F(bpm.T(4), dsp.Amplify(chord4, dsp.Sequence(dsp.Lerp(0, 1, bpm.T(2)), dsp.Lerp(1, 0, bpm.T(2))))),
	)
}
```
