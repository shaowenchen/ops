package kube

import "fmt"

func PrintError(errMsg string)(err error){
	fmt.Println(errMsg)
	return fmt.Errorf(errMsg)
}

func ErrorMsgGetClient(err error) string {
	return fmt.Sprintf("could not create kubernetes client: %v", err)
}

func ErrorMsgGetNode(err error) string {
	return fmt.Sprintf("could not get kubernetes node: %v", err)
}

func ErrorMsgGetNamespace(err error) string {
	return fmt.Sprintf("could not get kubernetes namespace: %v", err)
}

func ErrorMsgRunScriptOnNode(err error) string {
	return fmt.Sprintf("could not run script on node: %v", err)
}

func ErrorMsgLimitRange(err error) string {
	return fmt.Sprintf("could not change limitrange: %v", err)
}

func ErrorMsgImagePullSecret(err error) string {
	return fmt.Sprintf("could not change imagepullsecret: %v", err)
}

func ErrorMsgNodeName(err error) string {
	return fmt.Sprintf("could not change nodename: %v", err)
}

func ErrorMsgNodeSelector(err error) string {
	return fmt.Sprintf("could not change nodeselector: %v", err)
}