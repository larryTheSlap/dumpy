package subcmd

import (
	"dumpy/pkg/k8s"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	restartExample = `# restart capture mycap sniffers in current namespace
  kubectl dumpy restart mycap
# restart capture mycap sniffers in bar-ns with new tcpdump filters
  kubectl dumpy restart -f "-i any port 80"`
)

func Dumpysubcmd_restart() (cmd *cobra.Command) {

	dumpy := NewDumpy()

	cmd = &cobra.Command{
		Use:          "restart <captureName> [-n captureNamespace] [-f tcpdump filters]",
		Short:        "redeploy new sniffers for specified capture",
		Example:      restartExample,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := dumpy.Restart_Complete(); err != nil {
				return err
			}
			if err := dumpy.Restart_Validate(args, cmd); err != nil {
				return err
			}
			if err := dumpy.Restart_Run(args); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&dumpy.Namespace, "namespace", "n", "", "dumpy capture sniffers namespace (default: current namespace)")
	cmd.Flags().StringVarP(&dumpy.DumpFilters, "filter", "f", "", "tcpdump filters/arguments")
	return
}

func (d *Dumpy) Restart_Complete() (err error) {
	if err := d.Api.Set_ClientSet(); err != nil {
		return err
	}
	if d.Namespace == "" {
		d.Namespace, err = d.Api.Get_currentNS()
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Dumpy) Restart_Validate(args []string, cmd *cobra.Command) error {
	if len(args) == 0 {
		cmd.Usage()
		return errors.New("no argument provided")
	}
	if len(args) != 1 {
		return errors.New("unkown arguments, restart command require capture name. use -h for help")
	}
	return nil
}

func (d *Dumpy) Restart_Run(args []string) (err error) {
	d.CaptureName = args[0]
	var newDumpFilters string
	if d.DumpFilters != "" {
		newDumpFilters = d.DumpFilters
	}
	if err := d.NewSniffersFromExisting(); err != nil {
		return err
	}
	to_del := d.Sniffers
	if len(to_del) == 0 {
		return fmt.Errorf("%s sniffers not found in namespace %s", d.CaptureName, d.Namespace)
	}
	d.TargetResource, err = k8s.GetT_Resource(d.CaptureName, to_del[0].Namespace, d.Api)
	if err != nil {
		return err
	}
	fmt.Printf("performing restart operation on %s\n", d.CaptureName)
	if newDumpFilters != "" {
		d.DumpFilters = newDumpFilters
	}
	d.NewSniffers()
	fmt.Printf("deploying new %s sniffers...\n", d.CaptureName)
	if err := d.Sniff(); err != nil {
		return err
	}
	d.Sniffers = to_del
	fmt.Printf("\nRemoving old sniffers:\n")
	if err := d.Delete_Sniffers(); err != nil {
		return err
	}

	fmt.Printf("\n%s sniffers have been successfully redeployed\n", d.CaptureName)
	return nil
}
