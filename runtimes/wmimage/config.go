package wmimage

type Config struct {
	Seed         int64
	TargetSize   int
	BaseStrength float64
	MinEnergy    float64
	DWTLevels    int
	RSData       int
	RSPare       int
}

func DefaultConfig(seed int64) Config {
	return Config{
		Seed:         seed,
		TargetSize:   512,
		BaseStrength: 4.0,
		MinEnergy:    1200,
		DWTLevels:    2,
		RSData:       5,
		RSPare:       2,
	}
}
