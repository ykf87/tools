package ipquality

type IPInfo struct {
	IP        string
	Country   string
	City      string
	ASN       string
	Org       string
	IsHosting bool
}

type IPType string

const (
	IPTypeResidential IPType = "residential"
	IPTypeMobile      IPType = "mobile"
	IPTypeBusiness    IPType = "business"
	IPTypeHosting     IPType = "hosting"
	IPTypeUnknown     IPType = "unknown"
)

type QualityResult struct {
	Score int
	Type  IPType
	Tags  []string
	Raw   *IPInfo
}

type Decision struct {
	Allow  bool
	Reason string
	Score  int
	Type   IPType
}
