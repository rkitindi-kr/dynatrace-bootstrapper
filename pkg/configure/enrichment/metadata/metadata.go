package metadata

import (
	"path/filepath"

	"github.com/rkitindi-kr/dynatrace-bootstrapper/cmd/configure/attributes/container"
	"github.com/rkitindi-kr/dynatrace-bootstrapper/cmd/configure/attributes/pod"
	fsutils "github.com/rkitindi-kr/dynatrace-bootstrapper/pkg/utils/fs"
	"github.com/go-logr/logr"
	"github.com/spf13/afero"
)

const (
	JSONFilePath       = "enrichment/dt_metadata.json"
	PropertiesFilePath = "enrichment/dt_metadata.properties"
)

func Configure(log logr.Logger, fs afero.Afero, configDirectory string, podAttr pod.Attributes, containerAttr container.Attributes) error {
	confContent := fromAttributes(containerAttr, podAttr)

	log.V(1).Info("format content into a raw form", "struct", confContent)

	confJSON, err := confContent.toJSON()
	if err != nil {
		return err
	}

	jsonFilePath := filepath.Join(configDirectory, JSONFilePath)

	err = fsutils.CreateFile(fs, jsonFilePath, string(confJSON))
	if err != nil {
		log.Error(err, "failed to create metadata-enrichment properties file", "struct", jsonFilePath)

		return err
	}

	confProperties, err := confContent.toProperties()
	if err != nil {
		return err
	}

	propsFilePath := filepath.Join(configDirectory, PropertiesFilePath)

	err = fsutils.CreateFile(fs, propsFilePath, confProperties)
	if err != nil {
		log.Error(err, "failed to create metadata-enrichment properties file", "struct", propsFilePath)

		return err
	}

	return nil
}
