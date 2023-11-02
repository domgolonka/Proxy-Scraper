package providers

import "regexp"

var (
	portParamsRegexp = regexp.MustCompile(`([a-z]=\d;){10}`)
	portRegexp       = regexp.MustCompile(`(\+[a-z]){2,4}`)
	ipRegexp         = regexp.MustCompile(`(\d{1,3}\.){3}\d{1,3}`)
)
