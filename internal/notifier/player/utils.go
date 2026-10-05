package audio

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/flac"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/vorbis"
	"github.com/gopxl/beep/v2/wav"
)

func getStreamerAndFormat(f *os.File, filePath string) (beep.StreamCloser, beep.Format, error) {
	var streamer beep.StreamCloser
	var format beep.Format
	var err error

	switch strings.ToLower(filepath.Ext(filePath)) {
	case ".mp3":
		streamer, format, err = mp3.Decode(f)
	case ".wav":
		streamer, format, err = wav.Decode(f)
	case ".ogg":
		streamer, format, err = vorbis.Decode(f)
	case ".flac":
		streamer, format, err = flac.Decode(f)
	default:
		err = fmt.Errorf("unsupported audio format: %s", filePath)
	}
	if err != nil {
		return nil, beep.Format{}, err
	}
	return streamer, format, nil
}
