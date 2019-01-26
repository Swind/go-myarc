package main

import (
	"fmt"

	"github.com/gordonklaus/portaudio"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// RecordStreamHandler When MicrophoneDevice recording, use this handler to handle the data
type RecordStreamHandler interface {
	Open() error
	Close() error
	Write(buffer []int16) bool
}

// MicrophoneDevice the microphone device
type MicrophoneDevice struct {
	numInputChannels int
	sampleRate       float64
	framesPerBuffer  int
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

// NewMicrophoneDevice Create a new MicrophoneDevice
func NewMicrophoneDevice(numInputChannels int, sampleRate float64, framesPerBuffer int) *MicrophoneDevice {
	device := new(MicrophoneDevice)

	device.numInputChannels = numInputChannels
	device.sampleRate = sampleRate
	device.framesPerBuffer = framesPerBuffer

	return device
}

// Start start recording
func (device *MicrophoneDevice) Start(handler RecordStreamHandler) error {
	// Init portaudio
	zap.S().Debugw("Initializing portaudio ...")
	portaudio.Initialize()
	defer portaudio.Terminate()

	// Open handler
	handler.Open()
	defer handler.Close()

	// Open stream
	zap.S().Infow("Opening the default stream ...")
	buffer := make([]int16, device.framesPerBuffer)
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
	/*
		logger, _ := zap.NewDevelopment()
		zap.ReplaceGlobals(logger)
		logger.Sugar().Infow("An info message", "iteration", 1)

		numChannels := 2
		sampleRate := 44100
		framePerBuffer := 64

		handler := Handler.NewWavRecordStreamHandler(
			"test.wav",
			numChannels,
			sampleRate,
			framePerBuffer,
		)

		device := NewMicrophoneDevice(
			numChannels,
			float64(sampleRate),
			framePerBuffer,
		)

		err := device.Start(handler)
		if err != nil {
			panic(err)
		}
	*/
	example()
}
