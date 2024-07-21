package main

import (
	"os"

	"dumpy/pkg/subcmd"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	dumpyExample = `# capture all traffic from deployment foo-deploy with capture name mycap
  kubectl dumpy capture deploy foo-deploy --name mycap
# get mycap capture details
  kubectl dumpy get mycap
# export pcap files from mycap capture to target dir /tmp/dumps
  kubectl dumpy export mycap /tmp/dumps
# restart capture mycap sniffers
  kubectl dumpy restart mycap
# stop tpcdump capture on mycap sniffers
  kubectl dumpy stop mycap
# delete capture mycap 
  kubectl dumpy delete mycap
  `
)

func main() {
	flags := pflag.NewFlagSet("kubectl-dumpy", pflag.ExitOnError)
	pflag.CommandLine = flags
	if err := dumpyCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func dumpyCmd() (cmd *cobra.Command) {

	cmd = &cobra.Command{
		Use:          "dumpy [command]",
		Short:        "Perform network capture on containers running in a kubernetes cluster",
		Example:      dumpyExample,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Usage()
			return nil
		},
	}
	cmd.AddCommand(subcmd.Dumpysubcmd_capture())
	cmd.AddCommand(subcmd.Dumpysubcmd_export())
	cmd.AddCommand(subcmd.Dumpysubcmd_stop())
	cmd.AddCommand(subcmd.Dumpysubcmd_restart())
	cmd.AddCommand(subcmd.Dumpysubcmd_delete())
	cmd.AddCommand(subcmd.Dumpysubcmd_get())

	return cmd
}
