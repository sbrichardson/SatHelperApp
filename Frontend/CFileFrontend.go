package Frontend

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/OpenSatelliteProject/SatHelperApp/Logger"
	. "github.com/logrusorgru/aurora"
	"os"
	"time"
)

const CFileFrontendBufferSize = 65535

// region Struct Definition
type CFileFrontend struct {
	running         bool
	filename        string
	cb              SamplesCallback
	sampleRate      uint32
	centerFrequency uint32
	sampleBuffer    []complex64
	fileHandler     *os.File
	t0              time.Time
	fastAsPossible  bool
}

// endregion
// region Constructor
func NewCFileFrontend(filename string) *CFileFrontend {
	return &CFileFrontend{
		filename:        filename,
		running:         false,
		sampleRate:      0,
		centerFrequency: 0,
		cb:              nil,
		fastAsPossible:  false,
	}
}

// endregion
// region Getters
func (f *CFileFrontend) GetName() string {
	return fmt.Sprintf("CFileFrontend (%s)", f.filename)
}
func (f *CFileFrontend) GetShortName() string {
	return "CFileFrontend"
}
func (f *CFileFrontend) GetAvailableSampleRates() []uint32 {
	return make([]uint32, 0)
}
func (f *CFileFrontend) GetCenterFrequency() uint32 {
	return f.centerFrequency
}
func (f *CFileFrontend) GetSampleRate() uint32 {
	return f.sampleRate
}
func (f *CFileFrontend) EnableFastAsPossible() {
	f.fastAsPossible = true
}

// endregion
// region Setters
func (f *CFileFrontend) SetSamplesAvailableCallback(cb SamplesCallback) {
	f.cb = cb
}
func (f *CFileFrontend) SetSampleRate(sampleRate uint32) uint32 {
	f.sampleRate = sampleRate
	return sampleRate
}
func (f *CFileFrontend) SetCenterFrequency(centerFrequency uint32) uint32 {
	f.centerFrequency = centerFrequency
	return centerFrequency
}

// endregion
// region Commands
func (f *CFileFrontend) Init() bool {
	return true
}
func (f *CFileFrontend) Destroy() {}
func (f *CFileFrontend) Start() {
	if f.running {
		SLog.Error("CFileFrontend is already running.")
		return
	}
	f.running = true
	go func(frontend *CFileFrontend) {
		SLog.Info("CFileFrontend Routine started")
		f, err := os.Open(f.filename)

		var period = CFileFrontendBufferSize / float32(frontend.sampleRate)

		if err != nil {
			SLog.Error("Error opening file %s: %s", Bold(frontend.filename), Bold(err))
			frontend.running = false
			return
		}
		defer frontend.fileHandler.Close()

		frontend.fileHandler = f
		frontend.t0 = time.Now()
		frontend.sampleBuffer = make([]complex64, CFileFrontendBufferSize)

		var reader = bufio.NewReader(f)

		for frontend.running {
			if float32(time.Now().Sub(frontend.t0).Seconds()) >= period {
				err := binary.Read(reader, binary.LittleEndian, frontend.sampleBuffer)
				if err != nil {
					SLog.Error("Error reading input CFile: %s", Bold(err))
					frontend.running = false
					break
				}
				if frontend.cb != nil {
					var cbData = SampleCallbackData{
						SampleType:   SampleTypeFloatIQ,
						NumSamples:   len(frontend.sampleBuffer),
						ComplexArray: frontend.sampleBuffer,
					}
					frontend.cb(cbData)
				}
				frontend.t0 = time.Now()
			}
			if !frontend.fastAsPossible {
				time.Sleep(time.Duration((period / 100) * float32(time.Second)))
			}
		}
		SLog.Error("CFileFrontend Routine ended")
	}(f)
}

func (f *CFileFrontend) Stop() {
	if !f.running {
		SLog.Error("CFileFrontend is not running")
		return
	}
	f.running = false
}

func (f *CFileFrontend) SetAntenna(string) {}
func (f *CFileFrontend) SetAGC(bool)       {}
func (f *CFileFrontend) SetGain1(uint8)    {}
func (f *CFileFrontend) SetGain2(uint8)    {}
func (f *CFileFrontend) SetGain3(uint8)    {}
func (f *CFileFrontend) SetBiasT(bool)     {}

// endregion
