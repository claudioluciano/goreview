package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	lipgloss "charm.land/lipgloss/v2"
	"github.com/claudioluciano/goreview/internal/cli/styles"
	"github.com/claudioluciano/goreview/internal/platform"
	"github.com/spf13/cobra"
)

func newAuthCmd() *cobra.Command {
	var token string

	cmd := &cobra.Command{
		Use:   "auth <platform>",
		Short: "Configure authentication for a platform",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pt := platform.PlatformType(args[0])
			if pt != platform.GitHub && pt != platform.GitLab {
				return fmt.Errorf("unsupported platform: %s (use 'github' or 'gitlab')", args[0])
			}

			if token == "" {
				lipgloss.Print(styles.Prompt.Render("Enter token for "+string(pt)+": ") + " ")
				reader := bufio.NewReader(os.Stdin)
				input, err := reader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("read token: %w", err)
				}
				token = strings.TrimSpace(input)
			}

			if token == "" {
				return fmt.Errorf("token cannot be empty")
			}

			auth, err := platform.LoadAuth()
			if err != nil {
				return err
			}

			switch pt {
			case platform.GitHub:
				auth.GitHub = token
			case platform.GitLab:
				auth.GitLab = token
			}

			if err := platform.SaveAuth(auth); err != nil {
				return fmt.Errorf("save auth: %w", err)
			}

			lipgloss.Println(styles.Success.Render("Token saved for " + string(pt)))
			return nil
		},
	}

	cmd.Flags().StringVar(&token, "token", "", "Authentication token")

	return cmd
}
