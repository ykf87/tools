package ipquality

func DetectType(info *IPInfo, asnType ASNType) IPType {
	switch asnType {
	case ASNHosting:
		return IPTypeHosting
	case ASNMobile:
		return IPTypeMobile
	case ASNResidential:
		return IPTypeResidential
	default:
		return IPTypeBusiness
	}
}

func Score(info *IPInfo, t IPType) *QualityResult {
	score := 50

	switch t {
	case IPTypeResidential:
		score = 90
	case IPTypeMobile:
		score = 85
	case IPTypeBusiness:
		score = 60
	case IPTypeHosting:
		score = 20
	}

	return &QualityResult{
		Score: score,
		Type:  t,
		Tags:  []string{string(t)},
		Raw:   info,
	}
}
