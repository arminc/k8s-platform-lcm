package vulnerabilities

func filterAcceptedSeverities(vul []Vulnerability, severities []string) []Vulnerability {
	tmp := []Vulnerability{}
	for _, v := range vul {
		for _, severity := range severities {
			if v.Severity == severity {
				tmp = append(tmp, v)
			}
		}
	}
	return difference(vul, tmp)
}

func difference(a, b []Vulnerability) []Vulnerability {
	target := map[Vulnerability]bool{}
	for _, x := range b {
		target[x] = true
	}

	result := []Vulnerability{}
	for _, x := range a {
		if _, ok := target[x]; !ok {
			result = append(result, x)
		}
	}
	return result
}
