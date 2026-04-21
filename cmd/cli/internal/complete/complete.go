package complete

import (
	"strings"

	"github.com/spf13/cobra"
)

// Static returns candidates whose prefix matches toComplete.
func Static(candidates []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var out []string
	for _, c := range candidates {
		if strings.HasPrefix(c, toComplete) {
			out = append(out, c)
		}
	}
	return out, cobra.ShellCompDirectiveNoFileComp
}
