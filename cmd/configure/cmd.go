package configure

import (
	"path/filepath"

	"github.com/rkitindi-kr/dynatrace-bootstrapper/cmd/configure/attributes/container"
	"github.com/rkitindi-kr/dynatrace-bootstrapper/cmd/configure/attributes/pod"
	"github.com/rkitindi-kr/dynatrace-bootstrapper/pkg/configure/enrichment/endpoint"
	"github.com/rkitindi-kr/dynatrace-bootstrapper/pkg/configure/enrichment/metadata"
	"github.com/rkitindi-kr/dynatrace-bootstrapper/pkg/configure/oneagent/ca"
	"github.com/rkitindi-kr/dynatrace-bootstrapper/pkg/configure/oneagent/conf"
	"github.com/rkitindi-kr/dynatrace-bootstrapper/pkg/configure/oneagent/curl"
	"github.com/rkitindi-kr/dynatrace-bootstrapper/pkg/configure/oneagent/pmc"
	"github.com/rkitindi-kr/dynatrace-bootstrapper/pkg/configure/oneagent/preload"
	"github.com/go-logr/logr"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

const (
	InputFolderFlag  = "input-directory"
	ConfigFolderFlag = "config-directory"
	InstallPathFlag  = "install-path"
	IsFullstackFlag  = "fullstack"
	TenantFlag       = "tenant"
)

var (
	inputDir    string
	configDir   string
	installPath string
	isFullstack bool
	tenant      string

	podAttributes       []string
	containerAttributes []string
)

func AddFlags(cmd *cobra.Command) {
	// common
	cmd.PersistentFlags().StringVar(&inputDir, InputFolderFlag, "", "(Optional) Base path where to look for the configuration files.")
	cmd.PersistentFlags().StringVar(&configDir, ConfigFolderFlag, "", "(Optional) Base path where to put the configuration files.")
	cmd.PersistentFlags().StringArrayVar(&containerAttributes, container.Flag, []string{}, "(Optional) Container-specific attributes in JSON format.")
	cmd.PersistentFlags().StringArrayVar(&podAttributes, pod.Flag, []string{}, "(Optional) Pod-specific attributes in key=value format.")

	// oneagent
	cmd.PersistentFlags().StringVar(&installPath, InstallPathFlag, "/opt/dynatrace/oneagent", "(Optional) Base path where the agent binary will be put.")
	cmd.PersistentFlags().BoolVar(&isFullstack, IsFullstackFlag, false, "(Optional) Configure the CodeModule to be fullstack.")
	cmd.PersistentFlags().StringVar(&tenant, TenantFlag, "", "The name of the tenant that the CodeModule will communicate with. Mandatory in case of --fullstack.")

	cmd.PersistentFlags().Lookup(IsFullstackFlag).NoOptDefVal = "true"
}

func SetupOneAgent(log logr.Logger, fs afero.Afero, targetDir string) error {
	if configDir == "" || inputDir == "" {
		return nil
	}

	log.Info("starting configuration", "config-directory", configDir, "input-directory", inputDir)

	err := preload.Configure(log, fs, configDir, installPath)
	if err != nil {
		log.Info("failed to configure the ld.so.preload", "config-directory", configDir)

		return err
	}

	podAttr, err := pod.ParseAttributes(podAttributes)
	if err != nil {
		return err
	}

	containerAttrs, err := container.ParseAttributes(containerAttributes)
	if err != nil {
		return err
	}

	for _, containerAttr := range containerAttrs {
		containerConfigDir := filepath.Join(configDir, containerAttr.ContainerName)
		log.Info("starting to configure the container", "path", containerConfigDir)

		err = pmc.Configure(log, fs, inputDir, targetDir, containerConfigDir, installPath)
		if err != nil {
			log.Info("failed to configure the ruxitagentproc.conf", "config-directory", containerConfigDir)

			return err
		}

		err = conf.Configure(log, fs, containerConfigDir, containerAttr, podAttr, tenant, isFullstack)
		if err != nil {
			log.Info("failed to configure the container-conf files", "config-directory", containerConfigDir)

			return err
		}

		err = configureFromInputDir(log, fs, containerConfigDir, inputDir)
		if err != nil {
			log.Info("failed to configure container", "config-directory", containerConfigDir)

			return err
		}
	}

	log.Info("finished oneagent configuration", "config-directory", configDir, "input-directory", inputDir)

	return nil
}

func configureFromInputDir(log logr.Logger, fs afero.Afero, containerConfigDir, inputDir string) error {
	err := curl.Configure(log, fs, inputDir, containerConfigDir)
	if err != nil {
		log.Info("failed to configure the curl options", "config-directory", containerConfigDir)

		return err
	}

	err = ca.Configure(log, fs, inputDir, containerConfigDir)
	if err != nil {
		log.Info("failed to configure the CAs", "config-directory", containerConfigDir)

		return err
	}

	return nil
}

func EnrichWithMetadata(log logr.Logger, fs afero.Afero) error {
	if configDir == "" || inputDir == "" {
		return nil
	}

	log.Info("starting enrichment", "config-directory", configDir, "input-directory", inputDir)

	podAttr, err := pod.ParseAttributes(podAttributes)
	if err != nil {
		return err
	}

	containerAttrs, err := container.ParseAttributes(containerAttributes)
	if err != nil {
		return err
	}

	for _, containerAttr := range containerAttrs {
		containerConfigDir := filepath.Join(configDir, containerAttr.ContainerName)
		log.Info("starting to enrich the container", "path", containerConfigDir)

		err = endpoint.Configure(log, fs, inputDir, containerConfigDir)
		if err != nil {
			log.Info("failed to configure the endpoint.properties", "config-directory", configDir)

			return err
		}

		err = metadata.Configure(log, fs, containerConfigDir, podAttr, containerAttr)
		if err != nil {
			log.Info("failed to configure the enrichment files", "config-directory", containerConfigDir)

			return err
		}
	}

	log.Info("finished enrichment", "config-directory", configDir, "input-directory", inputDir)

	return nil
}
