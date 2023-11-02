package utils

import (
	"regexp"
	"strings"
)

func ParseIPs(body []byte, expression []string) []string {
	splitup := strings.Split(string(body), "\n")
	ipv4 := make([]string, 0, len(splitup))
	for i := 0; i < len(splitup); i++ {
		for _, exp := range expression {
			isIP := parseIps(splitup[i], exp)

			if isIP {
				ipv4 = append(ipv4, splitup[i])
			}
		}

	}
	return ipv4
}

// Check for IP:PORT
func parseIps(str string, expression string) bool {
	if str == "" {
		return false
	}
	ipv4WithPortCidrRegex := regexp.MustCompile(`` + expression + ``)
	return ipv4WithPortCidrRegex.MatchString(str)

}
