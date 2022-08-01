package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
)

var url = "https://raw.githubusercontent.com/shaowenchen/opscli/main/getopscli.sh"

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "upgrade opscli version to latest",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		httpClient := http.Client{
			Timeout: 3 * time.Second,
		}
		response, err := httpClient.Get(url)
		if err != nil {
			url = "https://ghproxy.com/" + url
			response, err = httpClient.Get(url)
			if err != nil {
				fmt.Print(err)
				return
			}
		}
		defer response.Body.Close()
		b, err := io.ReadAll(response.Body)
		if err != nil {
			return err
		}
		upgrade := exec.Command("sudo", "bash", "-c", string(b))
		_, err = upgrade.Output()
		if err != nil {
			fmt.Println("Upgrade failed!")
		}
		fmt.Println("Upgrade success!")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
