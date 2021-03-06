package env

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/ksonnet/ksonnet/metadata/app"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func stageFile(t *testing.T, fs afero.Fs, src, dest string) {
	in := filepath.Join("testdata", src)

	b, err := ioutil.ReadFile(in)
	require.NoError(t, err)

	dir := filepath.Dir(dest)
	err = fs.MkdirAll(dir, 0755)
	require.NoError(t, err)

	err = afero.WriteFile(fs, dest, b, 0644)
	require.NoError(t, err)
}

func withEnv(t *testing.T, fn func(afero.Fs)) {
	tmpDir, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// NOTE: using an os fs here because afero doesn't handle renames in the memmap version
	fs := afero.NewBasePathFs(afero.NewOsFs(), tmpDir)
	stageFile(t, fs, "app.yaml", "/app.yaml")

	dirs := []string{
		"env1",
		"env2",
		"nest/env3",
	}

	for _, dir := range dirs {
		path := filepath.Join("/", envRoot, dir)
		err := fs.MkdirAll(path, app.DefaultFolderPermissions)
		require.NoError(t, err)

		mainPath := filepath.Join(path, "main.jsonnet")
		stageFile(t, fs, "main.jsonnet", mainPath)

		paramsPath := filepath.Join(path, "params.libsonnet")
		stageFile(t, fs, "params.libsonnet", paramsPath)
	}

	componentParamsPath := filepath.Join("/", "components", "params.libsonnet")
	stageFile(t, fs, "component-params.libsonnet", componentParamsPath)

	fn(fs)
}

func checkExists(t *testing.T, fs afero.Fs, path string) {
	exists, err := afero.Exists(fs, path)
	require.NoError(t, err)

	require.True(t, exists, "%q should exist", path)
}

func checkNotExists(t *testing.T, fs afero.Fs, path string) {
	exists, err := afero.Exists(fs, path)
	require.NoError(t, err)

	require.False(t, exists, "%q should not exist", path)
}

func compareOutput(t *testing.T, fs afero.Fs, expected, got string) {
	gotData, err := afero.ReadFile(fs, got)
	require.NoError(t, err)

	expectedData, err := ioutil.ReadFile(filepath.Join("testdata", expected))
	require.NoError(t, err)

	require.Equal(t, string(expectedData), string(gotData))
}
