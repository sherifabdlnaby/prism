package validator

type imageType int

// Image types
const (
	JPEG imageType = iota
	PNG
	WEBP
)

func (_imageType imageType) String() string {
	name := []string{
		"JPEG",
		"PNG",
		"WEBP",
	}
	return name[_imageType]
}
