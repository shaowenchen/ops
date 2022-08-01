package script

import "fmt"

func AddHost(ip, domain string) string {
	return fmt.Sprintf(`sudo bash -c "echo '%s %s' >> /etc/hosts"`, ip, domain)
}

func DeleteHost(domain string) string {
	return fmt.Sprintf(`sudo bash -c "sed -i '/%s/d' /etc/hosts "`, domain)
}
