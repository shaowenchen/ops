package utils

import (
	"fmt"

	"net/http"
	"time"
)

func InstallOpscli() string {
	return fmt.Sprintf(`curl %s | sh -`, GetAvailableUrl("https://raw.githubusercontent.com/shaowenchen/opscli/main/getopscli.sh"))
}

func AddlMetricsServer() string {
	return fmt.Sprintf(`kubectl apply -f %s`, GetAvailableUrl("https://raw.githubusercontent.com/shaowenchen/image-syncer/main/kubernetes/metrics-server-0.5.0.yaml"))
}

func RemoveMetricsServer() string {
	return fmt.Sprintf(`kubectl delete -f %s`, GetAvailableUrl("https://raw.githubusercontent.com/shaowenchen/image-syncer/main/kubernetes/metrics-server-0.5.0.yaml"))
}

func AddHost(ip, domain string) string {
	return GetExecutableScript(fmt.Sprintf("echo '%s %s' >> /etc/hosts", ip, domain))
}

func DeleteHost(domain string) string {
	return GetExecutableScript(fmt.Sprintf("sed -i '/%s/d' /etc/hosts", domain))
}

func GetExecutableScript(script string) string {
	return fmt.Sprintf("sh -c '%s'", script)
}

func GetAvailableUrl(url string) string {
	httpClient := http.Client{
		Timeout: 3 * time.Second,
	}
	_, err := httpClient.Get(url)
	if err != nil {
		url = "https://ghproxy.com/" + url
	}
	return url
}
