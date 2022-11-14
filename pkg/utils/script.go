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
	return BuildBase64Cmd(sudo, fmt.Sprintf(`mv -bf %s %s`, src, dst))
}

func ScriptCopy(sudo bool, src string, dst string) string {
	return BuildBase64Cmd(sudo, fmt.Sprintf(`cp %s %s`, src, dst))
}

func ScriptMakeDir(sudo bool, src string) string {
	return BuildBase64Cmd(sudo, fmt.Sprintf(`mkdir -p %s`, src))
}

func ScriptChown(sudo bool, idU, idG, src string) string {
	return BuildBase64Cmd(sudo, fmt.Sprintf("chown %s:%s %s", idU, idG, src))
}

func ScriptRm(sudo bool, dst string) string {
	return BuildBase64Cmd(sudo, fmt.Sprintf(`rm -f %s`, dst))
}

func ScriptCPUTotal(sudo bool) string {
	return BuildBase64Cmd(sudo, `grep -c "model name" /proc/cpuinfo`)
}

func ScriptCPULoad1(sudo bool) string {
	return BuildBase64Cmd(sudo, `top -bn1 | grep load | awk '{printf "%.2f\n", $(NF-2)}'`)
}

func ScriptCPUUsagePercent(sudo bool) string {
	return BuildBase64Cmd(sudo, `grep 'cpu ' /proc/stat | awk '{usage=($2+$4)*100/($2+$4+$5)} END {printf ("%.2f%",usage)}'`)
}

func ScriptMemTotal(sudo bool) string {
	return BuildBase64Cmd(sudo, `free -h|grep -E "Mem"|awk '{print $2}'|head -1`)
}

func ScriptMemUsagePercent(sudo bool) string {
	return BuildBase64Cmd(sudo, `free -m | awk 'NR==2{printf "%.2f%\n", $3*100/$2 }'`)
}

func ScriptDiskTotal(sudo bool) string {
	return BuildBase64Cmd(sudo, `df -H | grep -vE '^Filesystem|tmpfs|cdrom|loop|udev' | awk '{ print $5 " " $2 " " $1 }' | grep " "/| head -n 1| awk '{ print $2}'`)
}

func ScriptDiskUsagePercent(sudo bool) string {
	return BuildBase64Cmd(sudo, `df -H | grep -vE '^Filesystem|tmpfs|cdrom|loop|udev' | awk '{ print $5 " " $2 " " $1 }' | grep " "/| head -n 1| awk '{ print $1}'`)
}

func ScriptHostname(sudo bool) string {
	return BuildBase64Cmd(sudo, `hostname`)
}

func ScriptKernelVersion(sudo bool) string {
	return BuildBase64Cmd(sudo, `uname -r`)
}

func ScriptDistribution(sudo bool) string {
	return BuildBase64Cmd(sudo, `cat /etc/os-release 2>/dev/null | grep ^ID= | awk -F= '{print $2}'`)
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
