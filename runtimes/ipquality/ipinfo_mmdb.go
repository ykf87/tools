package ipquality

import (
	"net"

	"github.com/oschwald/maxminddb-golang"
)

type IPInfoMMDB struct {
	db *maxminddb.Reader
}

func NewIPInfoMMDB(path string) (*IPInfoMMDB, error) {
	db, err := maxminddb.Open(path)
	if err != nil {
		return nil, err
	}
	return &IPInfoMMDB{db: db}, nil
}

type ipinfoRecord struct {
	Country string `maxminddb:"country"`
	ASN     string `maxminddb:"asn"`
	Org     string `maxminddb:"org"`
	Hosting bool   `maxminddb:"hosting"`
}

func (i *IPInfoMMDB) Lookup(ip string) (*IPInfo, error) {
	parsed := net.ParseIP(ip)

	var rec ipinfoRecord
	err := i.db.Lookup(parsed, &rec)
	if err != nil {
		return nil, err
	}

	return &IPInfo{
		IP:        ip,
		Country:   rec.Country,
		ASN:       rec.ASN,
		Org:       rec.Org,
		IsHosting: rec.Hosting,
	}, nil
}
