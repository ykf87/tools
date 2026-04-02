package ipquality

import (
	"net"

	"github.com/oschwald/geoip2-golang"
)

type GeoIP struct {
	city *geoip2.Reader
	asn  *geoip2.Reader
}

func NewGeoIP(cityPath, asnPath string) (*GeoIP, error) {
	city, err := geoip2.Open(cityPath)
	if err != nil {
		return nil, err
	}

	asn, err := geoip2.Open(asnPath)
	if err != nil {
		return nil, err
	}

	return &GeoIP{city: city, asn: asn}, nil
}

func (g *GeoIP) Lookup(ip string) (*IPInfo, error) {
	parsed := net.ParseIP(ip)

	city, _ := g.city.City(parsed)
	asn, _ := g.asn.ASN(parsed)

	return &IPInfo{
		IP:      ip,
		Country: city.Country.IsoCode,
		City:    city.City.Names["en"],
		ASN:     "AS" + itoa(asn.AutonomousSystemNumber),
		Org:     asn.AutonomousSystemOrganization,
	}, nil
}
