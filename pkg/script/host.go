package script

import (
	"fmt"

	"net/http"
	"time"
)

func InstallOpscli() string {
	return fmt.Sprintf(`curl %s | sh -`, GetAvailableUrl("https://raw.githubusercontent.com/shaowenchen/opscli/main/getopscli.sh"))
}

func InstallMetricsServer(clear bool) string {
	if clear {
		return fmt.Sprintf(`kubectl delete -f %s`, GetAvailableUrl("https://raw.githubusercontent.com/shaowenchen/image-syncer/main/kubernetes/metrics-server-0.5.0.yaml"))
	}
	return fmt.Sprintf(`kubectl apply -f %s`, GetAvailableUrl("https://raw.githubusercontent.com/shaowenchen/image-syncer/main/kubernetes/metrics-server-0.5.0.yaml"))
}
func AddHost(ip, domain string) string {
	return fmt.Sprintf(`sh -c "echo '%s %s' >> /etc/hosts"`, ip, domain)
}

func DeleteHost(domain string) string {
	return fmt.Sprintf(`sh -c "sed -i '/%s/d' /etc/hosts "`, domain)
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
