package audio

import (
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/speaker"
)

const outputSampleRate = beep.SampleRate(44100)

type AudioPlayer struct {
	buffer *beep.Buffer // decoded, resampled audio cached in memory
	format beep.Format

	speakerOnce sync.Once
	speakerErr  error

	mu      sync.Mutex
	stream  beep.StreamSeeker
	playing bool
}

func NewAudioPlayer(filePath string) (*AudioPlayer, error) {
	p := &AudioPlayer{}
	if err := p.load(filePath); err != nil {
		return nil, err
	}
	return p, nil
}

// load decodes any supported format (mp3, wav, ogg, flac) once, resamples
// it to a fixed rate, and buffers it fully in memory.
func (p *AudioPlayer) load(filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}

	defer func() {
		if err := f.Close(); err != nil {
			slog.Error("close file", "error", err)
		}
	}()

	streamer, format, err := getStreamerAndFormat(f, filePath)
	if err != nil {
		return fmt.Errorf("decode %s: %w", filePath, err)
	}

	defer func() {
		if err := streamer.Close(); err != nil {
			slog.Error("close streamer", "error", err)
		}
	}()

	resampled := beep.Resample(4, format.SampleRate, outputSampleRate, streamer)

	buf := beep.NewBuffer(beep.Format{
		SampleRate:  outputSampleRate,
		NumChannels: format.NumChannels,
		Precision:   format.Precision,
	})
	buf.Append(resampled)

	p.buffer = buf
	p.format = buf.Format()
	return nil
}

func (p *AudioPlayer) initSpeaker() {
	p.speakerErr = speaker.Init(outputSampleRate, outputSampleRate.N(time.Second/10))
}

// Play stops any current playback of this player, then plays from the
// start. Blocks until finished (or Stop is called).
func (p *AudioPlayer) Play() error {
	p.speakerOnce.Do(p.initSpeaker)
	if p.speakerErr != nil {
		return fmt.Errorf("init speaker: %w", p.speakerErr)
	}

	p.mu.Lock()
	speaker.Lock()
	if p.stream != nil {
		p.playing = false
	}
	newStream := p.buffer.Streamer(0, p.buffer.Len())
	p.stream = newStream
	p.playing = true
	speaker.Unlock()
	p.mu.Unlock()

	done := make(chan struct{})
	speaker.Play(beep.Seq(newStream, beep.Callback(func() {
		close(done)
	})))

	<-done

	p.mu.Lock()
	if p.stream == newStream {
		p.playing = false
		p.stream = nil
	}
	p.mu.Unlock()

	return nil
}

// Stop halts playback immediately if something is playing.
func (p *AudioPlayer) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.stream != nil {
		speaker.Lock()
		p.stream.Seek(p.stream.Len()) // jump to end -> stops output
		speaker.Unlock()
		p.stream = nil
		p.playing = false
	}
}

func (p *AudioPlayer) IsPlaying() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.playing
}
