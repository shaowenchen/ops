package host

import "fmt"

func PrintError(errMsg string) (err error) {
	fmt.Println(errMsg)
	return fmt.Errorf(errMsg)
}

func PrintInfo(infoMsg string){
	fmt.Println(infoMsg)
}

func ErrorCommon(err error) string {
	return fmt.Sprintf("err: %v", err)
}

func ErrorConnect(err error) string {
	return fmt.Sprintf("could not connect host: %v", err)
}

func ErrorEtcHosts(err error) string {
	return fmt.Sprintf("could not change /etc/hosts: %v", err)
}

func ErrorInstall(err error) string {
	return fmt.Sprintf("install component: %v", err)
}
