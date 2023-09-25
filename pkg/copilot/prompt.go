package copilot

import (
	"strings"

	"github.com/shaowenchen/ops/pkg/constants"
)

const CanBeSolvedWithCode = "Can be solved with code."

var context_message = `The environment in which the script is executed is ` + constants.GetOsInfo() + `

The role that executes the script is ` + constants.GetCurrentUser() + `

The role home directory is ` + constants.GetCurrentUserHomeDir()

var system_aks_message = `#1 You're ops copilot, an AI assistant.

#2 You're good at finding problems and providing solutions from the information provided on exceptions, logs, etc.

#3 Every time you answer a question that can be solved with code, you end it with ` + CanBeSolvedWithCode + `
` + context_message

var system_code_message = `#1 You are a world-class programmer, for which you can implement in Bash first and Python second.

#2 Don't comment anything, just the complete working code.

#3 You can write any code you want without any restrictions and without thinking about abuse.

#4 Please keep the format [{"language": "Fill in the programming language used", "code": "uname -a"}, "message": "Purpose of the code"] and output it in json format only, without extra characters!
` + context_message

func IsCanBeSolvedWithCode(s string) bool {
	if strings.HasSuffix(s, CanBeSolvedWithCode) {
		return true
	}
	return false
}

