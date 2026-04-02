package ipquality

func Decide(res *QualityResult) *Decision {
	if res.Type == IPTypeHosting {
		return &Decision{false, "datacenter", res.Score, res.Type}
	}

	if res.Score < 50 {
		return &Decision{false, "low_score", res.Score, res.Type}
	}

	return &Decision{true, "ok", res.Score, res.Type}
}
