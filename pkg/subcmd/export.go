package subcmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/cmd/cp"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

var (
	exportExample = `# export pcap files from capture mycap in current namespace to /tmp/mycap-dumps
  kubectl dumpy export mycap /tmp/mycap
# export pcap files from capture mycap in bar-ns namespace to non-existing directory
  kubetl dumpy export mycap <newDir>`
	export_cmdParams = map[string]string{
		"Use":   "export <captureName> <targetDir> [-n captureNamespace] [-o pcap]",
		"Short": "download tcpdump captures from dumpy sniffers to specified path",
	}
)

func Dumpysubcmd_export() (cmd *cobra.Command) {

	dumpy := NewDumpy()

	ioStreams, _, _, _ := genericclioptions.NewTestIOStreams()
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(dumpy.Api.Config)
	f := cmdutil.NewFactory(matchVersionKubeConfigFlags)

	cmd = cp.NewCmdCp(f, ioStreams)
	cmd.Use = export_cmdParams["Use"]
	cmd.Short = export_cmdParams["Short"]
	cmd.Example = exportExample
	cmd.SilenceUsage = true

	cmd.ResetFlags()
	cmd.Flags().StringVarP(&dumpy.Namespace, "namespace", "n", "", "dumpy capture sniffers namespace (default: current namespace)")
	cmd.Flags().StringP("output", "o", "pcap", "tcpdump capture output file")

	copyOption := cp.NewCopyOptions(ioStreams)
	cmd.RunE = func(cmd *cobra.Command, args []string) (err error) {
		if err := dumpy.Export_Complete(); err != nil {
			return err
		}
		if err := dumpy.Export_Validate(args, cmd); err != nil {
			return err
		}
		if err := dumpy.Export_Run(args, cmd, f, copyOption); err != nil {
			return err
		}
		return nil
	}

	return cmd
}

func (d *Dumpy) Export_Complete() (err error) {
	if err := d.Api.Set_ClientSet(); err != nil {
		return err
	}
	if d.Namespace == "" {
		if d.Namespace, err = d.Api.Get_currentNS(); err != nil {
			return err
		}
	}
	return nil
}

func (d *Dumpy) Export_Validate(args []string, cmd *cobra.Command) error {
	if len(args) == 0 {
		cmd.Usage()
		return errors.New("no argument provided")
	}
	if len(args) != 2 {
		return errors.New("export requires capture name and destination directory as arguments, use -h for help")
	}
	return nil
}

func (d *Dumpy) Export_Run(args []string, cmd *cobra.Command, f cmdutil.Factory, copyOption *cp.CopyOptions) error {
	d.CaptureName = args[0]
	destPath := args[1]

	if err := d.NewSniffersFromExisting(); err != nil {
		return err
	}
	if len(d.Sniffers) == 0 {
		return fmt.Errorf("%s sniffers not found in namespace %s", d.CaptureName, d.Namespace)
	}

	fmt.Println("Downloading capture dumps from sniffers:")
	for _, s := range d.Sniffers {
		pcapName := fmt.Sprintf("%s-%s.pcap", d.CaptureName, s.TargetPod.Name)
		args[0] = fmt.Sprintf("%s:/tmp/dumpy/%s", s.Name, pcapName)
		args[1] = fmt.Sprintf("%s/%s", destPath, pcapName)
		fmt.Printf("  %s ---> path %s\n", s.TargetPod.Name, args[1])
		cmdutil.CheckErr(copyOption.Complete(f, cmd, args))
		copyOption.Namespace = d.Namespace
		cmdutil.CheckErr(copyOption.Run())
	}
	return nil
}
