package main

import (
	"fmt"
	"os"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/gordonklaus/portaudio"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// RecordStreamHandler When MicrophoneDevice recording, use this handler to handle the data
type RecordStreamHandler interface {
	Open() error
	Close() error
	Write(buffer []int8) bool
}

// WavRecordStreamHandler save the sound file
type WavRecordStreamHandler struct {
	path        string
	bitDepth    int
	numChannels int
	sampleRate  int

	encoder       *wav.Encoder
	format        *audio.Format
	bufferAdapter *audio.IntBuffer
}

// NewWavRecordStreamHandler Create WavRecordStreamHandler
func NewWavRecordStreamHandler(path string, bitDepth int, numChannels int, sampleRate int) *WavRecordStreamHandler {
	handler := new(WavRecordStreamHandler)

	handler.path = path
	handler.bitDepth = bitDepth
	handler.numChannels = numChannels
	handler.sampleRate = sampleRate

	handler.format = &audio.Format{
		NumChannels: numChannels,
		SampleRate:  sampleRate,
	}

	handler.bufferAdapter = &audio.IntBuffer{
		Data:   nil,
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
		handler.bitDepth,
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

var counter = 0

// Write save the sound to file as wav
func (handler *WavRecordStreamHandler) Write(buffer []int8) bool {
	handler.bufferAdapter.Data = buffer
	handler.encoder.Write(handler.bufferAdapter)

	if counter > 10 {
		return true
	} else {
		return false
	}
}

// MicrophoneDevice the microphone device
type MicrophoneDevice struct {
	numInputChannels int
	sampleRate       float64
	bitDepth         int
	framesPerBuffer  int
}

// NewMicrophoneDevice Create a new MicrophoneDevice
func NewMicrophoneDevice(bitDepth int, numInputChannels int, sampleRate float64, framesPerBuffer int) *MicrophoneDevice {
	device := new(MicrophoneDevice)

	device.numInputChannels = numInputChannels
	device.sampleRate = sampleRate
	device.bitDepth = bitDepth
	device.framesPerBuffer = framesPerBuffer

	return device
}

// Start start recording
func (device *MicrophoneDevice) Start(handler RecordStreamHandler) error {
	zap.S().Infow("Initializing portaudio ...")

	// Init portaudio
	portaudio.Initialize()
	defer portaudio.Terminate()

	// Open handler
	handler.Open()
	defer handler.Close()

	// Open stream
	zap.S().Infow("Opening the default stream ...")
	buffer := make([]int8, device.framesPerBuffer*(device.bitDepth/8))
	stream, err := portaudio.OpenDefaultStream(
		device.numInputChannels,
		0,
		device.sampleRate,
		device.framesPerBuffer,
		buffer)
	if err != nil {
		fmt.Print(err)
		zap.S().Errorw("Can't open the default stream")
		return errors.Wrap(err, "OpenDefaultStream failed")
	}
	defer stream.Close()

	// Start recording
	err = stream.Start()
	if err != nil {
		return errors.Wrap(err, "Can't start the streaming")
	}
	defer stream.Stop()

	for {
		err = stream.Read()
		if err != nil {
			return errors.Wrap(err, "Can't read data from the stream")
		}
		if !handler.Write(buffer) {
			return nil
		}
	}
}

func main() {
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)
	logger.Sugar().Infow("An info message", "iteration", 1)

	bitDepth := 16
	numChannels := 2
	sampleRate := 44100

	handler := NewWavRecordStreamHandler(
		"test.wav",
		bitDepth,
		numChannels,
		sampleRate,
	)

	device := NewMicrophoneDevice(
		bitDepth,
		numChannels,
		float64(sampleRate),
		16,
	)

	device.Start(handler)
}
