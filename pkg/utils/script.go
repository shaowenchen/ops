package utils

import (
	"fmt"

	"net/http"
	"time"
)

func ScriptIsInChina() string {
	return `curl --connect-timeout 2 https://raw.githubusercontent.com/`
}

func ScriptInstallOpscli(proxy string) string {
	return fmt.Sprintf(`curl %s | sh -`, GetAvailableUrl("https://raw.githubusercontent.com/shaowenchen/ops/main/getcli.sh", proxy))
}

func ScriptAddHost(ip, domain string) string {
	return fmt.Sprintf(`echo \"%s %s\" >> /etc/hosts`, ip, domain)
}

func ScriptDeleteHost(domain string) string {
	return fmt.Sprintf(`sed -i '/%s/d' /etc/hosts`, domain)
}

func ScriptMv(src string, dst string) string {
	return fmt.Sprintf(`mv -bf %s %s`, src, dst)
}

func ScriptCopy(src string, dst string) string {
	return fmt.Sprintf(`cp %s %s`, src, dst)
}

func ScriptMakeDir(src string) string {
	return fmt.Sprintf(`mkdir -p %s`, src)
}

func ScriptChown(idU, idG, src string) string {
	return fmt.Sprintf(`chown %s:%s %s`, idU, idG, src)
}

func ScriptRm(dst string) string {
	return fmt.Sprintf(`rm -f %s`, dst)
}

func ScriptCPUTotal() string {
	return `grep -c "model name" /proc/cpuinfo`
}

func ScriptCPULoad1() string {
	return `top -bn1 | grep load | awk '{printf "%.2f\n", $(NF-2)}'`
}

func ScriptCPUUsagePercent() string {
	return `grep 'cpu ' /proc/stat | awk '{usage=($2+$4)*100/($2+$4+$5)} END {printf ("%.2f%",usage)}'`
}

func ScriptMemTotal() string {
	return `free -h|grep -E "Mem"|awk '{print $2}'|head -1`
}

func ScriptMemUsagePercent() string {
	return `free -m | awk 'NR==2{printf "%.2f%\n", $3*100/$2 }'`
}

func ScriptDiskTotal() string {
	return `df -H | grep "/$" | awk '{ print $5 " " $2 " " $1 }' | grep " "/| head -n 1| awk '{ print $2}'`
}

func ScriptDiskUsagePercent() string {
	return `df -H | grep "/$" | awk '{ print $5 " " $2 " " $1 }' | grep " "/| head -n 1| awk '{ print $1}'`
}

func ScriptHostname() string {
	return `hostname`
}

func ScriptKernelVersion() string {
	return `uname -r`
}

func ScriptDistribution() string {
	return `cat /etc/os-release 2>/dev/null | grep ^ID= | awk -F= '{print $2}' | sed 's/\"//g'`
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
