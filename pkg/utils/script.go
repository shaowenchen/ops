package utils

import (
	"fmt"
	"strings"

	"net/http"
	"time"
)

func ShellOpscliDownServer(api, aeskey, localfile, remotefile string) string {
	return fmt.Sprintf(`opscli file --direction "down" --api "%s" --aeskey "%s" --localfile "%s" --remotefile "%s"`, api, aeskey, localfile, remotefile)
}

func ShellOpscliUploadServer(api, aeskey, localfile, remotefile string) string {
	return fmt.Sprintf(`opscli file --direction "upload" --api "%s" --aeskey "%s" --localfile "%s" --remotefile "%s"`, api, aeskey, localfile, remotefile)
}

func ShellOpscliDownS3(region, endpoint, bucket, ak, sk, localfile, remotefile string) string {
	return fmt.Sprintf(`opscli file --direction "down" --region "%s" --endpoint "%s" --bucket "%s" --ak "%s" --sk "%s" --localfile "%s" --remotefile "s3://%s"`, region, endpoint, bucket, ak, sk, localfile, remotefile)
}

func ShellOpscliUploadS3(region, endpoint, bucket, ak, sk, localfile, remotefile string) string {
	return fmt.Sprintf(`opscli file --direction "upload" --region "%s" --endpoint "%s" --bucket "%s" --ak "%s" --sk "%s" --localfile "%s" --remotefile "s3://%s"`, region, endpoint, bucket, ak, sk, localfile, remotefile)
}

func ShellDownloadFile(proxy, sourceUrl, distPath string) string {
	return fmt.Sprintf(`curl -sfL %s -o %s`, GetAvailableUrl(sourceUrl, proxy), distPath)
}

func ShellIsInChina() string {
	return `curl --connect-timeout 2 https://raw.githubusercontent.com/`
}

func ShellInstallManifests(proxy string) string {
	proxy = formatProxy(proxy)
	return fmt.Sprintf(`curl -sfL %s | VERSION=latest PROXY=%s sh -`, GetAvailableUrl("https://raw.githubusercontent.com/shaowenchen/ops/main/getmanifests.sh", proxy), proxy)
}

func ShellInstallOpscli(proxy string) string {
	proxy = formatProxy(proxy)
	return fmt.Sprintf(`curl -sfL %s | VERSION=latest PROXY=%s sh -`, GetAvailableUrl("https://raw.githubusercontent.com/shaowenchen/ops/main/getcli.sh", proxy), proxy)
}

func ShellAddHost(ip, domain string) string {
	return fmt.Sprintf(`echo \"%s %s\" >> /etc/hosts`, ip, domain)
}

func ShellDeleteHost(domain string) string {
	return fmt.Sprintf(`sed -i '/%s/d' /etc/hosts`, domain)
}

func ShellMv(src string, dst string) string {
	return fmt.Sprintf(`mv -bf %s %s`, src, dst)
}

func ShellCopy(src string, dst string) string {
	return fmt.Sprintf(`cp %s %s`, src, dst)
}

func ShellMakeDir(src string) string {
	return fmt.Sprintf(`mkdir -p %s`, src)
}

func ShellChown(idU, idG, src string) string {
	return fmt.Sprintf(`chown %s:%s %s`, idU, idG, src)
}

func ShellRm(dst string) string {
	return fmt.Sprintf(`rm -f %s`, dst)
}

func ShellCPUTotal() string {
	return `grep -c "processor" /proc/cpuinfo`
}

func ShellCPULoad1() string {
	return `top -bn1 | grep load | awk '{printf "%.2f\n", $(NF-2)}'`
}

func ShellCPUUsagePercent() string {
	return `grep 'cpu ' /proc/stat | awk '{usage=($2+$4)*100/($2+$4+$5)} END {printf ("%.2f%%\n",usage)}'`
}

func ShellMemTotal() string {
	return `free -h|grep -E "Mem"|awk '{print $2}'|head -1`
}

func ShellMemUsagePercent() string {
	return `free -m | awk 'NR==2{printf "%.2f%%\n", $3*100/$2 }'`
}

func ShellDiskTotal(timeout int) string {
	return fmt.Sprintf(`timeout %d df -H 2>/dev/null | grep "^/" | grep -v "/dev/loop" | grep -v "/boot" | grep -v "/dev/longhorn" | awk '{ print $5 " " $2 " " $1 }' | grep " "/ | awk '{ print $2 }' | tr '\n' ' '`, timeout)
}

func ShellDiskUsagePercent(timeout int) string {
	return fmt.Sprintf(`timeout %d df -H 2>/dev/null | grep "^/" | grep -v "/dev/loop" | grep -v "/boot" | grep -v "/dev/longhorn" | awk '{ print $5 " " $2 " " $1 }' | grep " "/ | awk '{ print $1}' | tr '\n' ' '`, timeout)
}

func ShellHostname() string {
	return `hostname`
}

func ShellKernelVersion() string {
	return `uname -r`
}

func ShellArch() string {
	return `uname -m`
}

func ShellDistribution() string {
	return `cat /etc/os-release 2>/dev/null | grep ^ID= | awk -F= '{print $2}' | sed 's/\"//g'`
}

func ShellAcceleratorVendor() string {
	return `(lspci | grep -qi "accelerators: Huawei" && echo "Huawei") || (lspci | grep -qi "controller: NVIDIA" && echo "NVIDIA")`
}

func ShellAcceleratorModel() string {
	return `(npu-smi info -t board -i 0 -c 0 2>/dev/null | grep "Chip Name" | awk '{print $NF}' && nvidia-smi --query-gpu=name --format=csv,noheader 2>/dev/null | cut -d' ' -f2- | head -1) || echo ""`
}

func ShellAcceleratorCount() string {
	return `(npu_count=$(npu-smi info -l 2>/dev/null | grep -o 'Total Count\s*:\s*[0-9]\+' | awk '{print $NF}' | sed 's/ //g'); [ -n "$npu_count" ] && echo "$npu_count") || (nvidia_count=$(nvidia-smi -L 2>/dev/null | wc -l | awk '{print $1}'); [ "$nvidia_count" -gt 0 ] && echo "$nvidia_count") || echo ""`
}
func GetAvailableUrl(url string, proxy string) string {
	proxy = formatProxy(proxy)
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

func formatProxy(proxy string) string {
	if !strings.HasSuffix(proxy, "/") {
		proxy = proxy + "/"
	}
	return proxy
}
