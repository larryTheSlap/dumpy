package subcmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	deleteExample = `# delete dumpy capture mycap in current namespace
  kubectl dumpy delete mycap`
)

func Dumpysubcmd_delete() (cmd *cobra.Command) {

	dumpy := NewDumpy()

	cmd = &cobra.Command{
		Use:          "delete <captureName> [-n captureNamespace]",
		Short:        "delete specified capture sniffers",
		Example:      deleteExample,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := dumpy.Delete_Complete(); err != nil {
				return err
			}
			if err := dumpy.Delete_Validate(cmd, args); err != nil {
				return err
			}
			if err := dumpy.Delete_Run(args); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&dumpy.Namespace, "namespace", "n", "", "dumpy capture sniffers namespace (default: current namespace)")
	return
}

func (d *Dumpy) Delete_Complete() (err error) {
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

func (d *Dumpy) Delete_Validate(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		cmd.Usage()
		return errors.New("no argument provided")
	}
	if len(args) != 1 {
		return errors.New("unkown arguments, stop command require capture name. use -h for help")
	}
	return nil
}

func (d *Dumpy) Delete_Run(args []string) (err error) {
	d.CaptureName = args[0]
	if err := d.NewSniffersFromExisting(); err != nil {
		return err
	}
	if len(d.Sniffers) == 0 {
		return fmt.Errorf("%s sniffers not found in namespace %s", d.CaptureName, d.Namespace)
	}
	fmt.Printf("Deleting %s sniffers..\n", d.CaptureName)
	if err := d.Delete_Sniffers(); err != nil {
		return err
	}
	fmt.Printf("dumpy capture %s successfully deleted\n", d.CaptureName)
	return nil
}
