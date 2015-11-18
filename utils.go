package http_json_api

import (
	"strings"
)

func parseRefer(url string) (protocol string, domain string) {
	url = strings.TrimSpace(url)

	if len(url) > 0 {
		start0 := strings.Index(url, "://")
		url0 := url[start0+3 : len(url)]
		surls := strings.Split(url0, "/")

		if len(surls) > 0 {
			domain = surls[0]
		}

		protocol = url[0:start0]
	}

	return
}
