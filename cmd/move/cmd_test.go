package move

import (
	"path/filepath"
	"testing"

	impl "github.com/rkitindi-kr/dynatrace-bootstrapper/pkg/move"
	"github.com/go-logr/zapr"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

var testLog = zapr.NewLogger(zap.NewExample())

func TestExecute(t *testing.T) {
	sourceDir := "/source"
	targetDir := "/target"

	t.Run("simple copy", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}

		// Create source directory and files
		workDir := "/work"
		_ = fs.MkdirAll(sourceDir, 0755)
		_ = afero.WriteFile(fs, sourceDir+"/file1.txt", []byte("file1 content"), 0644)
		_ = afero.WriteFile(fs, sourceDir+"/file2.txt", []byte("file2 content"), 0644)
		_ = afero.WriteFile(fs, filepath.Join(sourceDir, impl.InstallerVersionFilePath), []byte("123"), 0644)

		workFolder = workDir

		err := Execute(testLog, fs, sourceDir, targetDir)
		require.NoError(t, err)

		// Check if the target directory and files exist
		exists, err := afero.DirExists(fs, targetDir)
		require.NoError(t, err)
		assert.True(t, exists)

		file1Exists, err := afero.Exists(fs, targetDir+"/file1.txt")
		require.NoError(t, err)
		assert.True(t, file1Exists)

		file2Exists, err := afero.Exists(fs, targetDir+"/file2.txt")
		require.NoError(t, err)
		assert.True(t, file2Exists)

		// Check the content of the copied files
		content, err := afero.ReadFile(fs, targetDir+"/file1.txt")
		require.NoError(t, err)
		assert.Equal(t, "file1 content", string(content))

		content, err = afero.ReadFile(fs, targetDir+"/file2.txt")
		require.NoError(t, err)
		assert.Equal(t, "file2 content", string(content))

		// Check the cleanup happened of the copied files
		exists, err = afero.DirExists(fs, workDir)
		require.NoError(t, err)
		assert.False(t, exists)
	})
	t.Run("execute with technology param", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}

		manifestContent := `{
			"version": "1.0",
			"technologies": {
				"java": {
					"x86": [
						{"path": "fileA1.txt", "version": "1.0", "md5": "abc123"},
						{"path": "agent/installer.version", "version": "1.0", "md5": "abc123"}
					]
				},
				"python": {
					"arm": [
						{"path": "fileA2.txt", "version": "1.0", "md5": "ghi789"}
					]
				}
			}
		}`

		technologyList := "java"
		_ = fs.MkdirAll(sourceDir, 0755)
		_ = afero.WriteFile(fs, sourceDir+"/manifest.json", []byte(manifestContent), 0644)
		_ = afero.WriteFile(fs, sourceDir+"/fileA1.txt", []byte("fileA1 content"), 0644)
		_ = afero.WriteFile(fs, sourceDir+"/fileA2.txt", []byte("fileA2 content"), 0644)
		_ = afero.WriteFile(fs, filepath.Join(sourceDir, impl.InstallerVersionFilePath), []byte("123"), 0644)

		technology = technologyList

		err := Execute(testLog, fs, sourceDir, targetDir)
		require.NoError(t, err)

		// Check if the target directory and files exist
		exists, err := afero.DirExists(fs, targetDir)
		require.NoError(t, err)
		assert.True(t, exists)

		file1Exists, err := afero.Exists(fs, targetDir+"/fileA1.txt")
		require.NoError(t, err)
		assert.True(t, file1Exists)

		file2Exists, err := afero.Exists(fs, targetDir+"/fileA2.txt")
		require.NoError(t, err)
		assert.False(t, file2Exists)

		// Check the content of the copied files
		content, err := afero.ReadFile(fs, targetDir+"/fileA1.txt")
		require.NoError(t, err)
		assert.Equal(t, "fileA1 content", string(content))

		content, err = afero.ReadFile(fs, targetDir+"/fileA2.txt")
		require.Error(t, err)
		assert.Empty(t, string(content))
	})
}
