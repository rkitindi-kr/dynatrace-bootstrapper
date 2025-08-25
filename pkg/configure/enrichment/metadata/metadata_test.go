package metadata

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/rkitindi-kr/dynatrace-bootstrapper/cmd/configure/attributes/container"
	"github.com/rkitindi-kr/dynatrace-bootstrapper/cmd/configure/attributes/pod"
	"github.com/go-logr/zapr"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

var testLog = zapr.NewLogger(zap.NewExample())

func TestConfigure(t *testing.T) {
	podAttr := pod.Attributes{
		UserDefined: map[string]string{
			"beep": "boop",
			"tip":  "top",
		},
		PodInfo: pod.PodInfo{
			PodName:       "podname",
			PodUID:        "poduid",
			NodeName:      "nodename",
			NamespaceName: "namespacename",
		},
		ClusterInfo: pod.ClusterInfo{
			ClusterUID:      "clusteruid",
			ClusterName:     "clustername",
			DTClusterEntity: "dtclusterentity",
		},
		WorkloadInfo: pod.WorkloadInfo{
			WorkloadKind: "workloadkind",
			WorkloadName: "workloadname",
		},
	}
	containerAttr := container.Attributes{
		ContainerName: "containername",
	}
	configDir := "path/conf"

	t.Run("success", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}

		err := Configure(testLog, fs, configDir, podAttr, containerAttr)
		require.NoError(t, err)

		expectedContent, err := fromAttributes(containerAttr, podAttr).toMap()
		require.NoError(t, err)

		jsonFilePath := filepath.Join(configDir, JSONFilePath)
		jsonContent, err := fs.ReadFile(jsonFilePath)
		require.NoError(t, err)

		for key, value := range expectedContent {
			assert.Contains(t, string(jsonContent), fmt.Sprintf("\"%s\":\"%s\"", key, value))
		}

		propsContent, err := fs.ReadFile(filepath.Join(configDir, PropertiesFilePath))
		require.NoError(t, err)

		for key, value := range expectedContent {
			assert.Contains(t, string(propsContent), key+"="+value)
		}
	})
}
