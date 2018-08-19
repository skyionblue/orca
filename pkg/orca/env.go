package orca

import (
	"fmt"
	"io"
	"log"
	"os"

	"orca/pkg/utils"

	"github.com/spf13/cobra"
)

type envCmd struct {
	chartsFile   string
	name         string
	override     []string
	packedValues []string
	set          []string
	kubeContext  string
	tls          bool
	helmTLSStore string
	museum       string
	createNS     bool

	out io.Writer
}

// NewGetEnvCmd represents the get env command
func NewGetEnvCmd(out io.Writer) *cobra.Command {
	e := &envCmd{out: out}

	cmd := &cobra.Command{
		Use:   "env",
		Short: "Get list of Helm releases in an environment (Kubernetes namespace)",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			if e.tls && e.helmTLSStore == "" {
				log.Fatal("TLS is set to true and HELM_TLS_STORE is not defined")
			}
			releases := utils.GetInstalledReleases(e.kubeContext, e.name, e.helmTLSStore, e.tls, true, true)

			fmt.Println("charts:")
			for _, r := range releases {
				fmt.Println("- name:", r.ChartName)
				fmt.Println("  vesrion:", r.ChartVersion)
			}
		},
	}

	f := cmd.Flags()

	f.StringVar(&e.name, "name", "", "name of environment (namespace) to get")
	f.StringVar(&e.kubeContext, "kube-context", "", "kubernetes context to get from")
	f.BoolVar(&e.tls, "tls", true, "should use communication over TLS")
	f.StringVar(&e.helmTLSStore, "helm-tls-store", os.Getenv("HELM_TLS_STORE"), "directory with TLS certs and keys")
	return cmd
}

// NewDeployEnvCmd represents the deploy env command
func NewDeployEnvCmd(out io.Writer) *cobra.Command {
	e := &envCmd{out: out}

	cmd := &cobra.Command{
		Use:   "env",
		Short: "Deploy a list of Helm charts to an environment (Kubernetes namespace)",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			if e.tls && e.helmTLSStore == "" {
				log.Fatal("TLS is set to true and HELM_TLS_STORE is not defined")
			}
			if e.createNS {
				utils.CreateNamespace(e.name, e.kubeContext)
			}

			desiredReleases := utils.ChartsYamlToStruct(e.chartsFile, e.name)
			desiredReleases = utils.OverrideReleases(desiredReleases, e.override)
			installedReleases := utils.GetInstalledReleases(e.kubeContext, e.name, e.helmTLSStore, e.tls, true, false)
			releasesToInstall := utils.GetReleasesDelta(desiredReleases, installedReleases)

			utils.AddRepository(e.museum, false)
			utils.UpdateRepository(e.museum, false)
			utils.DeployChartsFromMuseum(releasesToInstall, e.kubeContext, e.name, e.museum, e.helmTLSStore, e.tls, e.packedValues, e.set)

			installedReleases = utils.GetInstalledReleases(e.kubeContext, e.name, e.helmTLSStore, e.tls, true, false)
			releasesToDelete := utils.GetReleasesDelta(installedReleases, desiredReleases)

			utils.DeleteReleases(releasesToDelete, e.kubeContext, e.helmTLSStore, e.tls)
		},
	}

	f := cmd.Flags()

	f.StringVarP(&e.chartsFile, "charts-file", "c", "", "path to file with list of Helm charts to install")
	f.StringSliceVar(&e.override, "override", []string{}, "chart to override with different version (can specify multiple): chart=version")
	f.StringVar(&e.name, "name", "", "name of environment (namespace) to deploy to")
	f.StringVar(&e.museum, "museum", "", "chart museum instance (name=url)")
	f.StringVar(&e.kubeContext, "kube-context", "", "kubernetes context to deploy to")
	f.StringSliceVarP(&e.packedValues, "values", "f", []string{}, "values file to use (packaged within the chart)")
	f.StringSliceVarP(&e.set, "set", "s", []string{}, "set additional parameters")
	f.BoolVar(&e.tls, "tls", true, "should use communication over TLS")
	f.StringVar(&e.helmTLSStore, "helm-tls-store", os.Getenv("HELM_TLS_STORE"), "directory with TLS certs and keys")
	f.BoolVar(&e.createNS, "create-ns", false, "should create new namespace")

	return cmd
}

// NewDeleteEnvCmd represents the delete env command
func NewDeleteEnvCmd(out io.Writer) *cobra.Command {
	e := &envCmd{out: out}

	cmd := &cobra.Command{
		Use:   "env",
		Short: "Delete an environment (Kubernetes namespace) along with all Helm releases in it",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			if e.tls && e.helmTLSStore == "" {
				log.Fatal("TLS is set to true and HELM_TLS_STORE is not defined")
			}
			releases := utils.GetInstalledReleases(e.kubeContext, e.name, e.helmTLSStore, e.tls, true, true)
			utils.DeleteReleases(releases, e.kubeContext, e.helmTLSStore, e.tls)
			utils.DeleteNamespace(e.name, e.kubeContext)
		},
	}

	f := cmd.Flags()

	f.StringVar(&e.name, "name", "", "name of environment (namespace) to delete")
	f.StringVar(&e.kubeContext, "kube-context", "", "kubernetes context to delete in")
	f.BoolVar(&e.tls, "tls", true, "should use communication over TLS")
	f.StringVar(&e.helmTLSStore, "helm-tls-store", os.Getenv("HELM_TLS_STORE"), "directory with TLS certs and keys")

	return cmd
}
