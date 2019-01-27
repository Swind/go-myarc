package handler

import (
	"os"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

// WavRecordStreamHandler save the sound file
type WavRecordStreamHandler struct {
	path        string
	bitDepth    int
	numChannels int
	sampleRate  int

	encoder       *wav.Encoder
	format        *audio.Format
	bufferAdapter *audio.IntBuffer

	stop bool
}

// NewWavRecordStreamHandler Create WavRecordStreamHandler
func NewWavRecordStreamHandler(path string, numChannels int, sampleRate int, bufferLength int) *WavRecordStreamHandler {
	handler := new(WavRecordStreamHandler)

	handler.path = path
	handler.numChannels = numChannels
	handler.sampleRate = sampleRate
	handler.stop = false

	handler.format = &audio.Format{
		NumChannels: numChannels,
		SampleRate:  sampleRate,
	}

	handler.bufferAdapter = &audio.IntBuffer{
		Data:   make([]int, bufferLength),
		Format: handler.format,
	}

	return handler
}

// Open open the file
func (handler *WavRecordStreamHandler) Open() error {
	outFile, err := os.Create(handler.path)
	if err != nil {
		return err
	}

	encoder := wav.NewEncoder(
		outFile,
		handler.sampleRate,
		16,
		handler.numChannels,
		1)

	handler.encoder = encoder
	return nil
}

// Close close the file
func (handler *WavRecordStreamHandler) Close() error {
	if handler.encoder != nil {
		err := handler.encoder.Close()
		return err
	}
	return nil
}

// Write save the sound to file as wav
func (handler *WavRecordStreamHandler) Write(buffer []int16) bool {
	for i := 0; i < len(buffer); i++ {
		handler.bufferAdapter.Data[i] = int(buffer[i])
	}
	handler.encoder.Write(handler.bufferAdapter)

	return !handler.stop
}
