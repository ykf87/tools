package ipquality

type ASNType int

const (
	ASNResidential ASNType = iota
	ASNMobile
	ASNBusiness
	ASNHosting
)

var ASNData = map[string]ASNType{

	// ===== 🌐 云厂商 / 机房（重点拦截）=====
	"AS16509":  ASNHosting, // Amazon AWS
	"AS14618":  ASNHosting, // Amazon
	"AS15169":  ASNHosting, // Google
	"AS396982": ASNHosting, // Google Cloud
	"AS8075":   ASNHosting, // Microsoft Azure
	"AS8068":   ASNHosting,
	"AS14061":  ASNHosting, // DigitalOcean
	"AS63949":  ASNHosting, // Linode
	"AS20473":  ASNHosting, // Choopa / Vultr
	"AS9009":   ASNHosting, // M247
	"AS16276":  ASNHosting, // OVH
	"AS24940":  ASNHosting, // Hetzner
	"AS31898":  ASNHosting, // Oracle Cloud
	"AS45102":  ASNHosting, // Alibaba Cloud
	"AS37963":  ASNHosting, // Alibaba Cloud
	"AS132203": ASNHosting, // Tencent Cloud
	"AS45090":  ASNHosting, // Tencent

	// ===== 🇨🇳 中国 =====
	"AS4134": ASNResidential, // China Telecom
	"AS4809": ASNResidential,
	"AS4812": ASNResidential,

	"AS4837": ASNResidential, // China Unicom
	"AS9929": ASNResidential,

	"AS56046": ASNMobile, // China Mobile
	"AS9808":  ASNMobile,

	// ===== 🇯🇵 日本 =====
	"AS2516":  ASNResidential, // KDDI
	"AS17676": ASNResidential, // SoftBank Broadband
	"AS4713":  ASNResidential, // NTT

	"AS9605":  ASNMobile, // DOCOMO
	"AS17677": ASNMobile, // SoftBank Mobile

	// ===== 🇺🇸 美国 =====
	"AS7922":  ASNResidential, // Comcast
	"AS7018":  ASNResidential, // AT&T
	"AS20057": ASNResidential,
	"AS22773": ASNResidential, // Cox
	"AS6128":  ASNResidential,

	"AS21928": ASNMobile, // T-Mobile
	"AS22394": ASNMobile,
	"AS6167":  ASNMobile, // Verizon Wireless

	// ===== 🇪🇺 欧洲 =====
	"AS3320": ASNResidential, // Deutsche Telekom
	"AS3209": ASNResidential, // Vodafone
	"AS2856": ASNResidential, // BT UK

	"AS12576": ASNMobile, // Orange Mobile
	"AS13184": ASNMobile, // Telefonica Mobile
}
