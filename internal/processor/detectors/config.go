package detectors


type config struct {
	Export export
	minSize int
	maxSize int
	shiftFactor float64
	scaleFactor float64
	angle float64
	iouThreshold float64
	circleMarker bool
	cascadeFile string
}
type export struct {
	Format  string `validate:"oneof=jpg jpeg png"`
	Quality int    `validate:"min=1,max=100"`
}
func defaultConfig() *config {
	return &config{
		minSize: 20,
		maxSize:1000,
		shiftFactor: 0.1,
		scaleFactor: 1.1,
		angle: 0.0,
		iouThreshold: 0.2,
		circleMarker: false,
		cascadeFile: "./data/facefinder",
		Export: export{
			Format:  "jpeg",
			Quality: 85,
		},
	}
}
