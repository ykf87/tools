package ipquality

import "strings"

// type ASNType int

// const (
// 	ASNResidential ASNType = iota
// 	ASNMobile
// 	ASNBusiness
// 	ASNHosting
// )

var asnMap = map[string]ASNType{
	"AS16509": ASNHosting,
	"AS15169": ASNHosting,
}

var mobileASN = map[string]bool{
	"AS56046":           true, // China Mobile
	"ASdocomo":          true,
	"ASsoftbank_mobile": true,
}

var residentialASN = map[string]bool{
	"AS4134":  true, // China Telecom
	"AS4837":  true, // China Unicom
	"AS17676": true, // SoftBank Japan
	"AS2516":  true, // KDDI
	"AS3320":  true, // Deutsche Telekom
	"AS7922":  true, // Comcast
}

func ClassifyASN(asn, org string, hosting bool) ASNType {
	// 1️⃣ 机房优先
	if hosting {
		return ASNHosting
	}

	// 2️⃣ 精确匹配（最重要）
	if t, ok := ASNData[asn]; ok {
		return t
	}

	// 3️⃣ fallback（弱判断，避免误判 mobile）
	orgLower := strings.ToLower(org)

	if contains(orgLower, "cloud") ||
		contains(orgLower, "data center") {
		return ASNHosting
	}

	if contains(orgLower, "telecom") ||
		contains(orgLower, "broadband") ||
		contains(orgLower, "fiber") {
		return ASNResidential
	}

	// ⚠️ 注意：这里不要再用 mobile 关键词！

	return ASNBusiness
}
