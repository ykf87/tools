package browser

import (
	"fmt"
	"math/rand"
	"time"
)

type SSLFeature struct {
	Name  string
	Value string
}

type WebGL struct {
	Group  string
	Values []string
}

var BrowserSSLFeatures = []*SSLFeature{
	&SSLFeature{Name: "TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA", Value: "0xc00a"},
	&SSLFeature{Name: "TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA", Value: "0xc014"},
	&SSLFeature{Name: "TLS_RSA_WITH_AES_256_CBC_SHA", Value: "0x0035"},
	&SSLFeature{Name: "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA", Value: "0xc009"},
	&SSLFeature{Name: "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA", Value: "0xc013"},
	&SSLFeature{Name: "TLS_RSA_WITH_AES_128_CBC_SHA", Value: "0x002f"},
	&SSLFeature{Name: "TLS_RSA_WITH_3DES_EDE_CBC_SHA", Value: "0x000a"},
	&SSLFeature{Name: "TLS_RSA_WITH_AES_128_GCM_SHA256", Value: "0x009c"},
	&SSLFeature{Name: "TLS_RSA_WITH_AES_256_GCM_SHA384", Value: "0x009d"},
	&SSLFeature{Name: "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256", Value: "0xc02f"},
	&SSLFeature{Name: "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384", Value: "0xc030"},
	&SSLFeature{Name: "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256", Value: "0xc02b"},
	&SSLFeature{Name: "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384", Value: "0xc02c"},
	&SSLFeature{Name: "TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256", Value: "0xcca9"},
	&SSLFeature{Name: "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256", Value: "0xcca8"},
	&SSLFeature{Name: "TLS_PSK_WITH_AES_128_CBC_SHA", Value: "0x008c"},
	&SSLFeature{Name: "TLS_PSK_WITH_AES_256_CBC_SHA", Value: "0x008d"},
	&SSLFeature{Name: "TLS_ECDHE_PSK_WITH_AES_128_CBC_SHA", Value: "0xc035"},
	&SSLFeature{Name: "TLS_ECDHE_PSK_WITH_AES_256_CBC_SHA", Value: "0xc036"},
	&SSLFeature{Name: "TLS_ECDHE_PSK_WITH_CHACHA20_POLY1305_SHA256", Value: "0xccac"},
	&SSLFeature{Name: "TLS_AES_128_GCM_SHA256", Value: "0x1301"},
	&SSLFeature{Name: "TLS_AES_256_GCM_SHA384", Value: "0x1302"},
	&SSLFeature{Name: "TLS_CHACHA20_POLY1305_SHA256", Value: "0x1303"},
}
var BrowserWebGLs = []*WebGL{
	&WebGL{Group: "Google Inc. (Intel)", Values: []string{
		"ANGLE (Intel, Intel(R) HD Graphics 520 Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (Intel, Intel(R) HD Graphics 5300 Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (Intel, Intel(R) HD Graphics 620 Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (Intel, Intel(R) HD Graphics 620 Direct3D9Ex vs_3_0 ps_3_0)",
		"ANGLE (Intel, Intel(R) HD Graphics Direct3D11 vs_4_1 ps_4_1)",
		"ANGLE (Intel, Intel(R) HD Graphics Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (Intel, Intel(R) HD Graphics Family Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (Intel, Intel(R) UHD Graphics 620 Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (Intel, Intel(R) HD Graphics 4400 Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (Intel, Intel(R) UHD Graphics Direct3D11 vs_5_0 ps_5_0, D3D11-27.20.100.8935)",
		"ANGLE (Intel, Intel(R) UHD Graphics Direct3D11 vs_5_0 ps_5_0, D3D11-26.20.100.7870)",
		"ANGLE (Intel, Intel(R) UHD Graphics 620 Direct3D11 vs_5_0 ps_5_0, D3D11-27.20.100.8681)",
		"ANGLE (Intel, Intel(R) HD Graphics 630 Direct3D11 vs_5_0 ps_5_0, D3D11-27.20.100.8681)",
		"ANGLE (Intel, Intel(R) HD Graphics 530 Direct3D11 vs_5_0 ps_5_0, D3D11-27.20.100.9466)",
		"ANGLE (Intel, Intel(R) HD Graphics 5500 Direct3D11 vs_5_0 ps_5_0, D3D11-20.19.15.5126)",
		"ANGLE (Intel, Intel(R) HD Graphics 6000 Direct3D11 vs_5_0 ps_5_0, D3D11-20.19.15.5126)",
		"ANGLE (Intel, Intel(R) HD Graphics 610 Direct3D11 vs_5_0 ps_5_0, D3D11-27.20.100.9466)",
		"ANGLE (Intel, Intel(R) HD Graphics 630 Direct3D11 vs_5_0 ps_5_0, D3D11-27.20.100.9168)",
		"ANGLE (Intel, Intel(R) HD Graphics Direct3D11 vs_5_0 ps_5_0, D3D11-27.21.14.6589)",
		"ANGLE (Intel, Intel(R) UHD Graphics 620 Direct3D11 vs_5_0 ps_5_0, D3D11-27.20.100.9126)",
		"ANGLE (Intel, Mesa Intel(R) UHD Graphics 620 (KBL GT2), OpenGL 4.6 (Core Profile) Mesa 21.2.2)",
	}},
	&WebGL{Group: "Google Inc. (NVIDIA)", Values: []string{
		"ANGLE (NVIDIA, NVIDIA GeForce GTX 1050 Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (NVIDIA, NVIDIA GeForce GTX 1050 Ti Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (NVIDIA, NVIDIA GeForce GTX 1660 Ti Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (NVIDIA, NVIDIA GeForce RTX 2070 SUPER Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (NVIDIA, NVIDIA GeForce GTX 750 Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (NVIDIA, NVIDIA Quadro K600 Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (NVIDIA, NVIDIA Quadro M1000M Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (NVIDIA, NVIDIA GeForce GTX 750 Direct3D9Ex vs_3_0 ps_3_0)",
		"ANGLE (NVIDIA, NVIDIA GeForce GTX 760 Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (NVIDIA, NVIDIA GeForce GTX 750 Direct3D9Ex vs_3_0 ps_3_0)",
		"ANGLE (NVIDIA, NVIDIA GeForce GTX 750 Ti Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (NVIDIA, NVIDIA GeForce GTX 750 Ti Direct3D9Ex vs_3_0 ps_3_0)",
		"ANGLE (NVIDIA, NVIDIA GeForce GTX 760 Direct3D9Ex vs_3_0 ps_3_0)",
		"ANGLE (NVIDIA, NVIDIA GeForce GTX 770 Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (NVIDIA, NVIDIA GeForce GTX 780 Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (NVIDIA, NVIDIA GeForce GTX 850M Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (NVIDIA, NVIDIA GeForce GTX 850M Direct3D9Ex vs_3_0 ps_3_0)",
		"ANGLE (NVIDIA, NVIDIA GeForce GTX 860M Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (NVIDIA, NVIDIA GeForce GTX 950 Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (NVIDIA, NVIDIA GeForce GTX 950 Direct3D9Ex vs_3_0 ps_3_0)",
		"ANGLE (NVIDIA, NVIDIA GeForce GTX 950M Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (NVIDIA, NVIDIA GeForce GTX 950M Direct3D9Ex vs_3_0 ps_3_0)",
		"ANGLE (NVIDIA, NVIDIA GeForce GTX 960 Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (NVIDIA, NVIDIA GeForce GTX 960 Direct3D9Ex vs_3_0 ps_3_0)",
		"ANGLE (NVIDIA, NVIDIA GeForce GTX 960M Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (NVIDIA, NVIDIA GeForce GTX 960M Direct3D9Ex vs_3_0 ps_3_0)",
		"ANGLE (NVIDIA, NVIDIA GeForce GTX 970 Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (NVIDIA, NVIDIA GeForce GTX 970 Direct3D9Ex vs_3_0 ps_3_0)",
		"ANGLE (NVIDIA, NVIDIA GeForce GTX 980 Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (NVIDIA, NVIDIA GeForce GTX 980 Direct3D9Ex vs_3_0 ps_3_0)",
		"ANGLE (NVIDIA, NVIDIA GeForce GTX 980 Ti Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (NVIDIA, NVIDIA GeForce GTX 980M Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (NVIDIA, NVIDIA GeForce MX130 Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (NVIDIA, NVIDIA GeForce MX150 Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (NVIDIA, NVIDIA GeForce MX230 Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (NVIDIA, NVIDIA GeForce MX250 Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (NVIDIA, NVIDIA GeForce RTX 2060 Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (NVIDIA, NVIDIA GeForce RTX 2060 Direct3D9Ex vs_3_0 ps_3_0)",
		"ANGLE (NVIDIA, NVIDIA GeForce RTX 2060 SUPER Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (NVIDIA, NVIDIA GeForce RTX 2070 Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (NVIDIA, NVIDIA Quadro K620 Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (NVIDIA, NVIDIA Quadro FX 380 Direct3D11 vs_4_0 ps_4_0)",
		"ANGLE (NVIDIA, NVIDIA Quadro NVS 295 Direct3D11 vs_4_0 ps_4_0)",
		"ANGLE (NVIDIA, NVIDIA Quadro P1000 Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (NVIDIA, NVIDIA Quadro P2000 Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (NVIDIA, NVIDIA Quadro P400 Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (NVIDIA, NVIDIA Quadro P4000 Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (NVIDIA, NVIDIA Quadro P600 Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (NVIDIA, NVIDIA Quadro P620 Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (NVIDIA, NVIDIA, NVIDIA GeForce GTX 1070 Direct3D11 vs_5_0 ps_5_0, D3D11-27.21.14.6079)",
		"ANGLE (NVIDIA, NVIDIA, NVIDIA GeForce GTX 750 Ti Direct3D11 vs_5_0 ps_5_0, D3D11-10.18.13.6881)",
		"ANGLE (NVIDIA, NVIDIA, NVIDIA GeForce GTX 970 Direct3D11 vs_5_0 ps_5_0, D3D11-27.21.14.5671)",
		"ANGLE (NVIDIA, NVIDIA, NVIDIA GeForce GTX 750 Ti Direct3D11 vs_5_0 ps_5_0, D3D11-27.21.14.5671)",
		"ANGLE (NVIDIA, NVIDIA Corporation, GeForce GTX 1050 Ti/PCIe/SSE2, OpenGL 4.5.0 NVIDIA 460.73.01)",
		"ANGLE (NVIDIA, NVIDIA Corporation, GeForce GTX 1050 Ti/PCIe/SSE2, OpenGL 4.5.0 NVIDIA 460.80)",
		"ANGLE (NVIDIA, NVIDIA Corporation, GeForce GTX 1050/PCIe/SSE2, OpenGL 4.5 core)",
		"ANGLE (NVIDIA, NVIDIA Corporation, GeForce GTX 1060 6GB/PCIe/SSE2, OpenGL 4.5 core)",
		"ANGLE (NVIDIA, NVIDIA Corporation, GeForce GTX 1080 Ti/PCIe/SSE2, OpenGL 4.5 core)",
		"ANGLE (NVIDIA, NVIDIA Corporation, GeForce GTX 1650/PCIe/SSE2, OpenGL 4.5 core)",
		"ANGLE (NVIDIA, NVIDIA Corporation, GeForce GTX 650/PCIe/SSE2, OpenGL 4.5 core)",
		"ANGLE (NVIDIA, NVIDIA Corporation, GeForce GTX 750 Ti/PCIe/SSE2, OpenGL 4.5 core)",
		"ANGLE (NVIDIA, NVIDIA Corporation, GeForce GTX 860M/PCIe/SSE2, OpenGL 4.5 core)",
		"ANGLE (NVIDIA, NVIDIA Corporation, GeForce GTX 950M/PCIe/SSE2, OpenGL 4.5 core)",
		"ANGLE (NVIDIA, NVIDIA Corporation, GeForce MX150/PCIe/SSE2, OpenGL 4.5 core)",
		"ANGLE (NVIDIA, NVIDIA Corporation, GeForce RTX 2070/PCIe/SSE2, OpenGL 4.5 core)",
		"ANGLE (NVIDIA, NVIDIA Corporation, NVIDIA GeForce GTX 660/PCIe/SSE2, OpenGL 4.5.0 NVIDIA 470.57.02)",
		"ANGLE (NVIDIA, NVIDIA Corporation, NVIDIA GeForce RTX 2060 SUPER/PCIe/SSE2, OpenGL 4.5.0 NVIDIA 470.63.01)",
		"ANGLE (NVIDIA, NVIDIA, NVIDIA GeForce GTX 1050 Ti Direct3D9Ex vs_3_0 ps_3_0, nvd3dumx.dll-26.21.14.4250)",
		"ANGLE (NVIDIA, NVIDIA, NVIDIA GeForce GTX 1060 5GB Direct3D11 vs_5_0 ps_5_0, D3D11-30.0.14.7168)",
		"ANGLE (NVIDIA, NVIDIA, NVIDIA GeForce GTX 1060 6GB Direct3D11 vs_5_0 ps_5_0, D3D11-30.0.14.7212)",
		"ANGLE (NVIDIA, NVIDIA, NVIDIA GeForce GTX 1070 Ti Direct3D11 vs_5_0 ps_5_0, D3D11-27.21.14.6677)",
		"ANGLE (NVIDIA, NVIDIA, NVIDIA GeForce GTX 1080 Ti Direct3D11 vs_5_0 ps_5_0, D3D11-30.0.14.7111)",
		"ANGLE (NVIDIA, NVIDIA, NVIDIA GeForce GTX 1650 Direct3D11 vs_5_0 ps_5_0, D3D11-30.0.14.7212)",
		"ANGLE (NVIDIA, NVIDIA, NVIDIA GeForce GTX 1650 Ti Direct3D11 vs_5_0 ps_5_0, D3D11-30.0.14.7111)",
		"ANGLE (NVIDIA, NVIDIA, NVIDIA GeForce GTX 1660 SUPER Direct3D11 vs_5_0 ps_5_0, D3D11-30.0.14.7196)",
		"ANGLE (NVIDIA, NVIDIA, NVIDIA GeForce GTX 1660 Ti Direct3D11 vs_5_0 ps_5_0, D3D11-30.0.14.7196)",
	}},
	&WebGL{Group: "Google Inc. (AMD)", Values: []string{
		"ANGLE (AMD, AMD Radeon (TM) R9 370 Series Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (AMD, AMD Radeon HD 7700 Series Direct3D9Ex vs_3_0 ps_3_0)",
		"ANGLE (AMD, ATI Mobility Radeon HD 4330 Direct3D11 vs_4_1 ps_4_1)",
		"ANGLE (AMD, ATI Mobility Radeon HD 4500 Series Direct3D11 vs_4_1 ps_4_1)",
		"ANGLE (AMD, ATI Mobility Radeon HD 5000 Series Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (AMD, ATI Mobility Radeon HD 5400 Series Direct3D11 vs_5_0 ps_5_0)",
		"ANGLE (AMD, AMD, Radeon (TM) RX 470 Graphics Direct3D11 vs_5_0 ps_5_0, D3D11-27.20.1034.6)",
		"ANGLE (AMD, AMD, AMD Radeon(TM) Graphics Direct3D11 vs_5_0 ps_5_0, D3D11-27.20.14028.11002)",
		"ANGLE (AMD, AMD, AMD Radeon RX 5700 XT Direct3D11 vs_5_0 ps_5_0, D3D11-30.0.13025.1000)",
		"ANGLE (AMD, AMD, AMD Radeon RX 6900 XT Direct3D11 vs_5_0 ps_5_0, D3D11-30.0.13011.1004)",
		"ANGLE (AMD, AMD, AMD Radeon(TM) Graphics Direct3D11 vs_5_0 ps_5_0, D3D11-30.0.13002.23)",
	}},
}

var langMap = map[string]string{
	"ca-ES":      "Catalan",
	"prs-AF":     "Dari",
	"ps-AF":      "Pashto",
	"sq-AL":      "Albanian",
	"hy-AM":      "Armenian",
	"pt-PT":      "Portuguese",
	"en-US":      "English",
	"es-AR":      "Spanish",
	"de-AT":      "German",
	"ar-AE":      "Arabic",
	"sv-SE":      "Swedish",
	"fi-FI":      "Finnish",
	"az-Latn-AZ": "Azerbaijani",
	"az-Cyrl-AZ": "Azerbaijani",
	"bs-BA":      "Bosnian",
	"hr-BA":      "Croatian",
	"sr-BA":      "Serbian",
	"nl-NL":      "Dutch",
	"bn-BD":      "Bengali",
	"en-AU":      "English",
	"fr-FR":      "French",
	"bg-BG":      "Bulgarian",
	"ar-BH":      "Arabic",
	"quz-BO":     "Quechua",
	"fr-BE":      "French",
	"pt-BR":      "Portuguese",
	"tn-ZA":      "Tswana",
	"be-BY":      "Belarus",
	"en-BZ":      "English",
	"fr-CA":      "French",
	"ms-BN":      "Malay",
	"arn-CL":     "Mapdangan",
	"es-CO":      "Spanish",
	"es-CR":      "Spanish",
	"es-ES":      "Spanish",
	"el-GR":      "Greek",
	"tr-TR":      "Turkish",
	"cs-CZ":      "Czech",
	"de-DE":      "German",
	"ar-SA":      "Arabic",
	"da-DK":      "Danish",
	"es-DO":      "Spanish",
	"ar-DZ":      "Arabic",
	"quz-EC":     "Quechua",
	"et-EE":      "Estonian",
	"ar-EG":      "Arabic",
	"se-FI":      "Northern",
	"de-CH":      "German",
	"fo-FO":      "Faroese",
	"ka-GE":      "Georgian",
	"kl-GL":      "Greenland",
	"qut-GT":     "Keeche",
	"zh-HK":      "Chinese",
	"es-HN":      "Spanish",
	"hr-HR":      "Croatian",
	"hu-HU":      "Hungarian",
	"he-IL":      "Hebrew",
	"ga-IE":      "Irish",
	"id-ID":      "Indonesian",
	"en-GB":      "English",
	"hi-IN":      "Hindi",
	"is-IS":      "Island",
	"fa-IR":      "Persian",
	"ar-IQ":      "Arabic",
	"it-IT":      "Italian",
	"en-JM":      "English",
	"ar-JO":      "Arabic",
	"ja-JP":      "Japanese",
	"km-KH":      "Khmer",
	"ar-KW":      "Arabic",
	"de-LI":      "German",
	"si-LK":      "Sinhala",
	"ar-LB":      "Arabic",
	"lt-LT":      "Lithuanian",
	"de-LU":      "German",
	"lv-LV":      "Latvian",
	"ar-LY":      "Arabic",
	"ar-MA":      "Arabic",
	"fr-MC":      "French",
	"ro-RO":      "Romanian",
	"sr-Latn-ME": "Serbian",
	"mk-MK":      "Macedonian",
	"mt-MT":      "Maltese",
	"dv-MV":      "Dhivehi",
	"es-MX":      "Spanish",
	"ms-MY":      "Malay",
	"ha-Latn-NG": "Hausa",
	"es-NI":      "Spain",
	"fy-NL":      "Frisian",
	"se-NO":      "Northern",
	"ne-NP":      "Nepal",
	"ar-OM":      "Arabic",
	"es-PA":      "Spanish",
	"quz-PE":     "Quechua",
	"zh-MO":      "Chinese",
	"fil-PH":     "Philippine",
	"pl-PL":      "Polish",
	"ur-PK":      "Urdu",
	"es-PR":      "Spain",
	"ar-QA":      "Arabic",
	"sr-Latn-RS": "Serbian",
	"es-PY":      "Spanish",
	"ru-RU":      "Russian",
	"se-SE":      "Northern",
	"en-SG":      "English",
	"sl-SI":      "Slovenian",
	"nn-no":      "Norwegian",
	"sk-SK":      "Slovak",
	"wo-SN":      "Wolof",
	"es-SV":      "Spanish",
	"ar-SY":      "Arabic",
	"th-TH":      "Thai",
	"ar-TN":      "Arabic",
	"sw-KE":      "Swah",
	"uk-UA":      "Ukrainian",
	"es-UY":      "Spanish",
	"es-VE":      "Spanish",
	"vi-VN":      "Vietnamese",
	"ar-YE":      "Arabic",
	"en-ZW":      "English",
	"nso-ZA":     "Basotho",
	"zh-CN":      "Chinese",
	"am-ET":      "Amharic",
	"en-TT":      "English",
	"zh-TW":      "Chinese",
	"en-NZ":      "English",
	"mi-NZ":      "Maori",
	"ko-KR":      "North",
	"lo-LA":      "Lao",
	"ky-KG":      "Kyrgyz",
	"kk-KZ":      "Kazakh",
	"tg-Cyrl-TJ": "Tajik",
	"tk-TM":      "Turkmen",
	"uz-Latn-UZ": "Uzbek",
	"mn-Mong":    "Mongolian",
	"rw-RW":      "Rwanda",
}

var BrowserMemorys = []int{2, 4, 8, 16, 32, 64}
var BrowserCpu = []int{2, 4, 6, 8, 12}

type AudioContextStruct struct {
	Analyer float64 `json:"analyer"`
	Channel float64 `json:"channel"`
	Mode    int     `json:"mode"`
}

func (this *AudioContextStruct) Random() {
	this.Analyer = rand.Float64() * 0.01
	this.Channel = rand.Float64() * 1e-6
	this.Mode = 1
}

type CanvasStruct struct {
	R    int `json:"r"`
	G    int `json:"g"`
	B    int `json:"b"`
	A    int `json:"a"`
	Mode int `json:"mode"`
}

func (this *CanvasStruct) Random() {
	this.A = rand.Intn(10)
	this.R = rand.Intn(21) - 10
	this.G = rand.Intn(21) - 10
	this.B = rand.Intn(21) - 10
	this.Mode = 1
}

type ClientRectsStruct struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	Mode   int     `json:"mode"`
}

func (this *ClientRectsStruct) Random() {
	this.Width = rand.Float64()*2 - 1
	this.Height = rand.Float64()*2 - 1
	this.Mode = 1
}

type CookieS struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Domain   string `json:"domain"`
	Path     string `json:"path"`
	Session  bool   `json:"session"`
	HttpOnly bool   `json:"httpOnly"`
	Secure   bool   `json:"secure"`
	SameSite string `json:"sameSite"`
}

type CookieStruct struct {
	JsonStr string     `json:"jsonStr"`
	Mode    int        `json:"mode"`
	Value   []*CookieS `json:"value"`
}

type LocationStruct struct {
	Enable    int     `json:"enable"` // 是否开启位置信息.1询问地址,2允许,3关闭
	Mode      int     `json:"mode"`   // 1自己设置位置信息,2根据ip自动,一般选2就行.如果是1,则必须设置lat和log
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Precision int     `json:"precision"`
}

type DeviceNameStruct struct {
	Mode  int    `json:"mode"`
	Value string `json:"value"`
}

func (this *DeviceNameStruct) Random() {
	rand.Seed(time.Now().UnixNano())
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	length := 8 // 生成 8 位随机字符

	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	this.Value = fmt.Sprintf("DESKTOP-%s", string(result))
	this.Mode = 1
}

type PortStruct struct {
	API      string `json:"API"`
	Host     string `json:"host"`
	Mode     int    `json:"mode"`
	Pass     string `json:"pass"`
	Port     string `json:"port"`
	Protocol string `json:"protocol"`
	User     string `json:"user"`
	Value    string `json:"value"`
}
type SecChUaStruct struct {
	Brand   string `json:"brand"`
	Version any    `json:"version"`
}

type SpeechVoicesStruct struct {
	Default      bool   `json:"default"`
	Lang         string `json:"lang"`
	LocalService bool   `json:"localService"`
	Name         string `json:"name"`
	VoiceURI     string `json:"voiceURI"`
}

type User struct {
	AudioContext  *AudioContextStruct `json:"audio-context"`
	Canvas        *CanvasStruct       `json:"canvas"`
	ChromeVersion string              `json:"chrome_version"`
	ClientRects   *ClientRectsStruct  `json:"ClientRects"`
	Cookie        []*CookieStruct     `json:"cookie"`
	Cpu           struct {
		Mode  int `json:"mode"`
		Value int `json:"value"`
	} `json:"cpu"`
	DeviceName *DeviceNameStruct `json:"device-name"`
	Dnt        struct {
		Mode  int `json:"mode"`
		Value int `json:"value"`
	} `json:"dnt"`
	Fonts struct {
		Mode  int      `json:"mode"`
		Value []string `json:"value"`
	} `json:"fonts"`
	Gpu struct {
		Mode  int `json:"mode"`
		Value int `json:"value"`
	} `json:"gpu"`
	Group    string `json:"group"`
	Homepage struct {
		Mode  int    `json:"mode"`
		Value string `json:"value"`
	} `json:"homepage"`
	Id        int64           `json:"id"`
	IsRunning bool            `json:"isRunning"`
	Location  *LocationStruct `json:"location"`
	Mac       struct {
		Mode  int    `json:"mode"`
		Value string `json:"value"`
	} `json:"mac"`
	Media struct {
		Mode int `json:"mode"`
	} `json:"media"`
	Memory struct {
		Mode  int `json:"mode"`
		Value int `json:"value"`
	}
	Name     string `json:"name"`
	Os       string `json:"os"`
	PortScan struct {
		Mode  int      `json:"mode"`
		Value []string `json:"value"`
	} `json:"port-scan"`
	Proxy  *PortStruct `json:"proxy"`
	Screen struct {
		Width  int    `json:"width"`
		Height int    `json:"height"`
		Mode   int    `json:"mode"`
		Value  string `json:"_value"`
	}
	SecChUa struct {
		Mode  int `json:"mode"`
		Value []SecChUaStruct
	} `json:"sec-ch-ua"`
	SpeechVoices struct {
		Mode  int                  `json:"mode"`
		Value []SpeechVoicesStruct `json:"value"`
	} `json:"speech_voices"`
}
