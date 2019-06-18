package validator

type config struct {
	Format    []string
	MaxHeight int `mapstructure:"max_height" validate:"min=0" `
	MaxWidth  int `mapstructure:"max_width" validate:"min=0"`
	MinHeight int `mapstructure:"min_height" validate:"min=0"`
	MinWidth  int `mapstructure:"min_width" validate:"min=0"`

	jpeg, png, webp bool
}

//defaultConfig returns the default configs
func defaultConfig() *config {
	return &config{}
}
