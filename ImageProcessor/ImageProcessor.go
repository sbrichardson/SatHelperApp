package ImageProcessor

import (
	"github.com/opensatelliteproject/SatHelperApp/ImageProcessor/ImageTools"
	"github.com/opensatelliteproject/SatHelperApp/ImageProcessor/MapCutter"
	"github.com/opensatelliteproject/SatHelperApp/ImageProcessor/MapDrawer"
	"github.com/opensatelliteproject/SatHelperApp/ImageProcessor/Structs"
	"github.com/opensatelliteproject/SatHelperApp/Logger"
	"github.com/opensatelliteproject/SatHelperApp/XRIT"
	"github.com/opensatelliteproject/SatHelperApp/XRIT/NOAAProductID"
	"github.com/opensatelliteproject/SatHelperApp/XRIT/PacketData"
	"sync"
)

var purgeFiles = false

type ImageProcessor struct {
	sync.Mutex
	MultiSegmentCache map[string]*Structs.MultiSegmentImage
	mapDrawer         *MapDrawer.MapDrawer
	mapCutter         *MapCutter.MapCutter
	reproject         bool
	drawmap           bool
	falsecolor        bool
	enhance           bool
	metadata          bool
	cutRegions        []string
}

func MakeImageProcessor() *ImageProcessor {
	return &ImageProcessor{
		MultiSegmentCache: make(map[string]*Structs.MultiSegmentImage),
		mapDrawer:         ImageTools.GetDefaultMapDrawer(),
		mapCutter:         ImageTools.GetDefaultMapCutter(),
		reproject:         false,
		drawmap:           false,
		falsecolor:        false,
		enhance:           false,
		metadata:          false,
		cutRegions:        make([]string, 0),
	}
}

func (ip *ImageProcessor) SetFalseColor(fsclr bool) {
	ip.falsecolor = fsclr
	if fsclr {
		SLog.Warn("False color is enabled, so it will also save plain images with no map")
		ImageTools.SetSaveNoMap(true) // Needed for FSCLR
	}
}

func (ip *ImageProcessor) SetCutRegions(regions []string) {
	ip.cutRegions = regions
}

func (ip *ImageProcessor) SetDrawMap(drawMap bool) {
	ip.drawmap = drawMap
}

func (ip *ImageProcessor) SetReproject(reproject bool) {
	ip.reproject = reproject
}

func (ip *ImageProcessor) SetMetadata(metadata bool) {
	ip.metadata = metadata
}

func (ip *ImageProcessor) SetEnhance(enhance bool) {
	ip.enhance = enhance
}

func (ip *ImageProcessor) GetFalseColor() bool {
	return ip.falsecolor
}

func (ip *ImageProcessor) GetDrawMap() bool {
	return ip.drawmap
}

func (ip *ImageProcessor) GetReproject() bool {
	return ip.reproject
}

func (ip *ImageProcessor) GetEnhance() bool {
	return ip.enhance
}

func (ip *ImageProcessor) GetMetadata() bool {
	return ip.metadata
}

func (ip *ImageProcessor) GetCutRegions() []string {
	return ip.cutRegions
}

func (ip *ImageProcessor) GetMapCutter() *MapCutter.MapCutter {
	return ip.mapCutter
}

func (ip *ImageProcessor) GetMapDrawer() *MapDrawer.MapDrawer {
	if ip.drawmap {
		return ip.mapDrawer
	}

	return nil
}

func (ip *ImageProcessor) ProcessImage(filename string) {
	ip.Lock()
	defer ip.Unlock()

	xh, err := XRIT.ParseFile(filename)
	if err != nil {
		SLog.Error("Error parsing file %s: %s", filename, err)
		return
	}

	if xh.PrimaryHeader.FileTypeCode != PacketData.IMAGE {
		return
	}

	switch xh.NOAASpecificHeader.ProductID {
	case NOAAProductID.GOES16_ABI, NOAAProductID.GOES17_ABI:
		ProcessGOESABI(ip, filename, xh)
	}

	ip.checkExpired()
}

func (ip *ImageProcessor) checkExpired() {
	for k, v := range ip.MultiSegmentCache {
		if v.Expired() {
			SLog.Warn("Image %s timed out waiting segments. Removing from cache.", k)
			delete(ip.MultiSegmentCache, k)
			if purgeFiles {
				v.Purge()
			}
		}
	}
}

func SetPurgeFiles(purge bool) {
	purgeFiles = purge
	SLog.Info("Set Purge Files changed to %v", purge)
}
