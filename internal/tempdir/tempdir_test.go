package tempdir

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	pb "gitlab.com/gitlab-org/gitaly-proto/go"
	"gitlab.com/gitlab-org/gitaly/internal/testhelper"

	"github.com/stretchr/testify/require"
)

func TestNewSuccess(t *testing.T) {
	ctx, cancel := testhelper.Context()
	defer cancel()
	repo := testhelper.TestRepository()

	tempDir, err := New(ctx, repo)
	require.NoError(t, err)

	err = ioutil.WriteFile(path.Join(tempDir, "test"), []byte("hello"), 0644)
	require.NoError(t, err, "write file in tempdir")

	cancel() // This should trigger async removal of the temporary directory

	// Poll because the directory removal is async
	for i := 0; i < 100; i++ {
		_, err = os.Stat(tempDir)
		if err != nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	require.True(t, os.IsNotExist(err), "expected directory to have been removed, got error %v", err)
}

func TestNewFailStorageUnknown(t *testing.T) {
	ctx, cancel := testhelper.Context()
	defer cancel()
	_, err := New(ctx, &pb.Repository{StorageName: "does-not-exist", RelativePath: "foobar.git"})
	require.Error(t, err)
}

var cleanRoot = path.Join("testdata/clean", tmpRootPrefix)

func TestCleanerSafety(t *testing.T) {
	defer func() {
		if p := recover(); p != nil {
			if _, ok := p.(invalidCleanRoot); !ok {
				t.Fatalf("expected invalidCleanRoot panic, got %v", p)
			}
		}
	}()

	//This directory is invalid because it does not end in '+gitaly/tmp'
	invalidDir := "testdata/does-not-exist"
	clean(invalidDir)

	t.Fatal("expected panic")
}

func TestCleanSuccess(t *testing.T) {
	if err := chmod("a", 0700); err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}

	require.NoError(t, os.RemoveAll(cleanRoot), "clean up test clean root")

	old := time.Unix(0, 0)
	recent := time.Now()

	makeDir(t, "a", old)
	makeDir(t, "c", recent)
	makeDir(t, "f", old)

	makeFile(t, "a/b", old)
	makeFile(t, "c/d", old)
	makeFile(t, "e", recent)

	// This is really evil and even breaks 'rm -rf'
	require.NoError(t, chmod("a", 0), "apply evil permissions to 'a'")

	assertEntries(t, "a", "c", "e", "f")

	require.NoError(t, clean(cleanRoot), "walk first pass")
	// 'a' won't get removed because it's mtime is bumped when 'a/b' is deleted
	assertEntries(t, "a", "c", "e")

	info, err := stat("a")
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0700), info.Mode().Perm(), "permissions of 'a' should have been fixed")

	_, err = stat("a/b")
	require.True(t, os.IsNotExist(err), "entry 'a/b' should be gone")

	require.NoError(t, clean(cleanRoot), "walk second pass")
	assertEntries(t, "a", "c", "e")
}

func chmod(p string, mode os.FileMode) error {
	return os.Chmod(path.Join(cleanRoot, p), mode)
}

func stat(p string) (os.FileInfo, error) {
	return os.Stat(path.Join(cleanRoot, p))
}

func assertEntries(t *testing.T, entries ...string) {
	foundEntries, err := ioutil.ReadDir(cleanRoot)
	require.NoError(t, err)

	require.Len(t, foundEntries, len(entries))

	for i, name := range entries {
		require.Equal(t, name, foundEntries[i].Name())
	}
}

func makeFile(t *testing.T, filePath string, mtime time.Time) {
	fullPath := path.Join(cleanRoot, filePath)
	require.NoError(t, ioutil.WriteFile(fullPath, nil, 0644))
	require.NoError(t, os.Chtimes(fullPath, mtime, mtime))
}

func makeDir(t *testing.T, dirPath string, mtime time.Time) {
	fullPath := path.Join(cleanRoot, dirPath)
	require.NoError(t, os.MkdirAll(fullPath, 0700))
	require.NoError(t, os.Chtimes(fullPath, mtime, mtime))
}

func TestCleanNoTmpExists(t *testing.T) {
	// This directory is valid because it ends in the special prefix
	dir := path.Join("testdata", "does-not-exist", tmpRootPrefix)

	require.NoError(t, clean(dir))
}
