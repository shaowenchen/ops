package utils

import (
	"fmt"

	"net/http"
	"time"
)

func ScriptInstallOpscli(proxy string) string {
	return fmt.Sprintf(`curl %s | sh -`, GetAvailableUrl("https://raw.githubusercontent.com/shaowenchen/ops/main/getcli.sh", proxy))
}

func ScriptAddHost(sudo bool, ip, domain string) string {
	return BuildBase64Cmd(sudo, fmt.Sprintf("echo \"%s %s\" >> /etc/hosts", ip, domain))
}

func ScriptDeleteHost(sudo bool, domain string) string {
	return BuildBase64Cmd(sudo, fmt.Sprintf("sed -i '/%s/d' /etc/hosts", domain))
}

func ScriptMv(sudo bool, src string, dst string) string {
	return fmt.Sprintf(`%s mv -bf %s %s`, GetSudoString(sudo), GetAbsoluteFilePath(src), GetAbsoluteFilePath(dst))
}

func ScriptCopy(sudo bool, src string, dst string) string {
	return fmt.Sprintf(`%s cp %s %s`, GetSudoString(sudo), GetAbsoluteFilePath(src), GetAbsoluteFilePath(dst))
}

func ScriptRm(sudo bool, dst string) string {
	return fmt.Sprintf(`%s rm -f %s %s`, GetSudoString(sudo), GetAbsoluteFilePath(dst))
}

func GetAvailableUrl(url string, proxy string) string {
	if proxy != "" {
		return proxy + url
	}
	httpClient := http.Client{
		Timeout: 3 * time.Second,
	}
	_, err := httpClient.Get(url)
	if err != nil {
		url = proxy + url
	}
	return url
}
