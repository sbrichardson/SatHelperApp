package Demuxer

import (
	"github.com/logrusorgru/aurora"
	"github.com/opensatelliteproject/SatHelperApp/Logger"
	"github.com/opensatelliteproject/SatHelperApp/ccsds"
)

type DirectDemuxer struct {
	demux *ccsds.Demuxer
}

func MakeDirectDemuxer(outFolder, tmpFolder string, drawMap, reproject, falseColor, metaframe, enhance bool) *DirectDemuxer {
	d := &DirectDemuxer{}

	d.demux = ccsds.MakeDemuxer()
	d.demux.SetOutputFolder(outFolder)
	d.demux.SetTemporaryFolder(tmpFolder)
	d.demux.SetDrawMap(drawMap)
	d.demux.SetReprojectImage(reproject)
	d.demux.SetFalseColor(falseColor)
	d.demux.SetMetaFrame(metaframe)
	d.demux.SetEnhance(enhance)

	SLog.Info("Starting direct Demuxer with: ")
	SLog.Info(" Output Folder: %s", aurora.Bold(outFolder).Green())
	SLog.Info(" Temporary Folder: %s", aurora.Bold(tmpFolder).Green())

	d.demux.SetOnFrameLost(func(channelId, currentFrame, lastFrame int) {
		SLog.Info("Lost Frames for channel %d: %d", channelId, currentFrame-lastFrame-1)
	})

	d.demux.SetOnNewVCID(func(channelId int) {
		SLog.Debug("New Channel: %d", channelId)
	})
	return d
}

func (dd *DirectDemuxer) AddSkipVCID(vcid int) {
	SLog.Info("Adding VCID %d to skip list.", vcid)
	dd.demux.AddSkipVCID(vcid)
}

func (dd *DirectDemuxer) Init()  {}
func (dd *DirectDemuxer) Start() {}
func (dd *DirectDemuxer) Stop()  {}
func (dd *DirectDemuxer) SendData(data []byte) {
	dd.demux.WriteBytes(data)
}
func (dd *DirectDemuxer) GetName() string {
	return "Direct"
}
