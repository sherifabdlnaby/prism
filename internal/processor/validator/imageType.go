package validator

type imageType int

const (
	JPEG imageType = iota
	PNG
)

func (_imageType imageType) String() string {
	name := []string{
		"JPEG",
		"PNG",
	}
	return name[_imageType]
}
