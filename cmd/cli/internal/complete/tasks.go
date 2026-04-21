package complete

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/shaowenchen/ops/pkg/constants"
	"github.com/spf13/cobra"
)

// OpsTaskYAMLBasenames returns task path completions under ~/.ops/tasks matching toComplete:
//   - stem without extension (e.g. get-os-usage), same as GetTaskAbsoluteFilePath resolution
//   - full basename (e.g. get-os-usage.yaml)
//   - tilde path (e.g. ~/.ops/tasks/get-os-usage.yaml) when the task dir is under $HOME
func OpsTaskYAMLBasenames(toComplete string) ([]string, cobra.ShellCompDirective) {
	dir := constants.GetOpsTaskDir()
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	home, errHome := os.UserHomeDir()
	taskDirUnderHome := errHome == nil && strings.HasPrefix(dir, home)

	var suffixFromHome string
	if taskDirUnderHome {
		suffixFromHome = strings.TrimPrefix(strings.TrimPrefix(dir, home), string(filepath.Separator))
	}

	seen := make(map[string]struct{})
	var raw []string
	add := func(s string) {
		if s == "" {
			return
		}
		if !strings.HasPrefix(s, toComplete) {
			return
		}
		if _, ok := seen[s]; ok {
			return
		}
		seen[s] = struct{}{}
		raw = append(raw, s)
	}

	for _, ent := range ents {
		if ent.IsDir() {
			continue
		}
		name := ent.Name()
		if !isYAMLTaskFile(name) {
			continue
		}
		ext := filepath.Ext(name)
		stem := strings.TrimSuffix(name, ext)
		add(stem)
		add(name)
		if taskDirUnderHome {
			add(filepath.Join("~", suffixFromHome, name))
		}
	}
	sort.Strings(raw)
	return raw, cobra.ShellCompDirectiveNoFileComp
}

func isYAMLTaskFile(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".yaml", ".yml":
		return true
	default:
		return false
	}
}

// TaskRunValidArgs completes --filepath / -f with basenames from ~/.ops/tasks (for commands with DisableFlagParsing).
func TaskRunValidArgs(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if strings.HasPrefix(toComplete, "--filepath=") {
		v := strings.TrimPrefix(toComplete, "--filepath=")
		comps, d := OpsTaskYAMLBasenames(v)
		for i := range comps {
			comps[i] = "--filepath=" + comps[i]
		}
		return comps, d
	}
	if strings.HasPrefix(toComplete, "-f=") {
		v := strings.TrimPrefix(toComplete, "-f=")
		comps, d := OpsTaskYAMLBasenames(v)
		for i := range comps {
			comps[i] = "-f=" + comps[i]
		}
		return comps, d
	}
	if len(args) == 0 {
		return nil, cobra.ShellCompDirectiveDefault
	}
	last := args[len(args)-1]
	if last == "--filepath" || last == "-f" {
		return OpsTaskYAMLBasenames(toComplete)
	}
	return nil, cobra.ShellCompDirectiveDefault
}

// FlagOpsTaskYAMLBasenames is for RegisterFlagCompletionFunc on normally-parsed commands.
func FlagOpsTaskYAMLBasenames(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return OpsTaskYAMLBasenames(toComplete)
}
