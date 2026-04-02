package ipquality

type Client struct {
	geo  *GeoIP
	ipdb *IPInfoMMDB
	mem  *MemoryCache
	sql  *SQLiteCache
}

func NewClient(geo *GeoIP, ipdb *IPInfoMMDB, sql *SQLiteCache) *Client {
	return &Client{
		geo:  geo,
		ipdb: ipdb,
		mem:  NewMemoryCache(100000),
		sql:  sql,
	}
}

func (c *Client) Check(ip string) (*Decision, error) {

	// 1. memory cache
	if v, ok := c.mem.Get(ip); ok {
		return v, nil
	}

	// 2. sqlite cache
	if v, ok := c.sql.Get(ip); ok {
		c.mem.Set(ip, v)
		return v, nil
	}

	// 3. lookup
	geo, _ := c.geo.Lookup(ip)
	info, _ := c.ipdb.Lookup(ip)

	merged := merge(geo, info)

	// 4. classify
	asnType := ClassifyASN(merged.ASN, merged.Org, merged.IsHosting)

	// 5. type
	t := DetectType(merged, asnType)

	// 6. score
	res := Score(merged, t)

	// 7. decision
	decision := Decide(res)

	// 8. cache
	c.mem.Set(ip, decision)
	c.sql.Set(ip, decision)

	return decision, nil
}
