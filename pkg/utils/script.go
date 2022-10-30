package utils

import (
	"fmt"

	"net/http"
	"strings"
	"time"
)

func ScriptInstallOpscli() string {
	return fmt.Sprintf(`curl %s | sh -`, GetAvailableUrl("https://raw.githubusercontent.com/shaowenchen/opscli/main/getopscli.sh"))
}

func ScriptInstallMetricsServer(clear string) string {
	if strings.ToLower(clear) == "true" {
		return fmt.Sprintf(`kubectl delete -f %s`, GetAvailableUrl("https://raw.githubusercontent.com/shaowenchen/image-syncer/main/kubernetes/metrics-server-0.5.0.yaml"))
	}
	return fmt.Sprintf(`kubectl apply -f %s`, GetAvailableUrl("https://raw.githubusercontent.com/shaowenchen/image-syncer/main/kubernetes/metrics-server-0.5.0.yaml"))
}

func ScriptAddHost(ip, domain string) string {
	return BuildBase64Cmd(fmt.Sprintf("echo \"%s %s\" >> /etc/hosts", ip, domain))
}

func ScriptDeleteHost(domain string) string {
	return BuildBase64Cmd(fmt.Sprintf("sed -i '/%s/d' /etc/hosts", domain))
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

func ClearOpscliPipelibeBackupDir() string {
	return "rm -rf ~/.opscli/.pipeline*"
}

func ClearOpscliLogDir() string {
	return "rm -rf ~/.opscli/logs/*"
}
