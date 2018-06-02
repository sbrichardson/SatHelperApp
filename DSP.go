package main

import (
	"log"
	"github.com/OpenSatelliteProject/libsathelper"
	"github.com/OpenSatelliteProject/SatHelperApp/Frontend"
	. "github.com/logrusorgru/aurora"
)

func initDSP() {
	circuitSampleRate := float32(device.GetSampleRate()) / float32(CurrentConfig.Base.Decimation)
	sps := circuitSampleRate / float32(CurrentConfig.Base.SymbolRate)

	log.Printf(Cyan("Samples per Symbol: %f").String(), Bold(Green(sps)))
	log.Printf(Cyan("Circuit Sample Rate: %f").String(), Bold(Green(circuitSampleRate)))
	log.Printf(Cyan("Low Pass Decimator Cut Frequency: %f").String(), Bold(Green(circuitSampleRate / 2)))

	rrcTaps := SatHelper.FiltersRRC(1, float64(circuitSampleRate), float64(CurrentConfig.Base.SymbolRate), float64(CurrentConfig.Base.RRCAlpha), RrcTaps)
	decimatorTaps := SatHelper.FiltersLowPass(1, float64(device.GetSampleRate()), float64(circuitSampleRate / 2), 100e3, SatHelper.FFTWindowsHAMMING, 6.76)

	decimator = SatHelper.NewFirFilter(uint(CurrentConfig.Base.Decimation), decimatorTaps)
	agc = SatHelper.NewAGC(AgcRate, AgcReference, AgcGain, AgcMaxGain)
	costasLoop = SatHelper.NewCostasLoop(PllAlpha, LoopOrder)
	clockRecovery = SatHelper.NewClockRecovery(sps, ClockGainOmega, ClockMu, ClockAlpha, ClockOmegaLimit)
	rrcFilter = SatHelper.NewFirFilter(1, rrcTaps)


	log.Printf(Cyan("Center Frequency: %d MHz").String(), Bold(Green(device.GetCenterFrequency())))
	log.Printf(Cyan("Automatic Gain Control: %t").String(), Bold(Green(CurrentConfig.Base.AGCEnabled)))

	if CurrentConfig.Base.AGCEnabled {
		device.SetAGC(true)
	} else {
		device.SetAGC(false)
		// TODO: Gains
	}
}

func newSamplesCallback(d Frontend.SampleCallbackData) {
	switch d.SampleType {
	case Frontend.FRONTEND_SAMPLETYPE_FLOATIQ: AddToFifoC64(samplesFifo, d.ComplexArray, d.NumSamples); break
	case Frontend.FRONTEND_SAMPLETYPE_S16IQ: AddToFifoS16(samplesFifo, d.Int16Array, d.NumSamples); break
	case Frontend.FRONTEND_SAMPLETYPE_S8IQ: AddToFifoS8(samplesFifo, d.Int8Array, d.NumSamples); break
	}
}

func processSamples() {
	if samplesFifo.Len() <= 64 * 1024{
		return
	}

	length := samplesFifo.Len()
	checkAndResizeBuffers(length)

	for i := 0; i < length; i++ {
		buffer0[i] = samplesFifo.Next().(complex64)
	}

	ba := &buffer0[0]
	bb := &buffer1[0]

	if CurrentConfig.Base.Decimation > 1 {
		length /= int(CurrentConfig.Base.Decimation)
		decimator.Work(ba, bb, length)
		swapBuffers(&ba, &bb)
	}

	agc.Work(ba, bb, length)
	swapBuffers(&ba, &bb)

	rrcFilter.Work(ba, bb, length)
	swapBuffers(&ba, &bb)

	costasLoop.Work(ba, bb, length)
	swapBuffers(&ba, &bb)

	symbols := clockRecovery.Work(ba, bb, length)
	swapBuffers(&ba, &bb)

	sendbuffer := make([]byte, symbols)

	var ob *[]complex64

	if ba == &buffer0[0] {
		ob = &buffer0
	} else {
		ob = &buffer1
	}

	for i := 0; i < symbols; i++ {
		z := (*ob)[i]
		v := imag(z) * 127
		if v > 127 {
			v = 127
		} else if v < -128 {
			v = -128
		}

		sendbuffer[i] = byte(v)
	}

	_, err := conn.Write(sendbuffer)
	if err != nil {
		log.Printf(Red("Error writting data: %s").String(), Bold(err))
	}

}