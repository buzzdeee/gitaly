package repository

import (
	"testing"

	"github.com/stretchr/testify/require"
	pb "gitlab.com/gitlab-org/gitaly-proto/go"
	"gitlab.com/gitlab-org/gitaly/internal/testhelper"
)

func TestWriteConfigSuccessful(t *testing.T) {
	ctx, cancel := testhelper.Context()
	defer cancel()

	server, serverSocketPath := runRepoServer(t)
	defer server.Stop()

	client, conn := newRepositoryClient(t, serverSocketPath)
	defer conn.Close()

	testRepo, _, cleanupFn := testhelper.NewTestRepo(t)
	defer cleanupFn()

	c, err := client.WriteConfig(ctx, &pb.WriteConfigRequest{Repository: testRepo, FullPath: "foo/bar"})
	require.NoError(t, err)
	require.NotNil(t, c)
	require.Empty(t, c.GetError())
}

func TestWriteConfigFailure(t *testing.T) {
	ctx, cancel := testhelper.Context()
	defer cancel()

	server, serverSocketPath := runRepoServer(t)
	defer server.Stop()

	client, conn := newRepositoryClient(t, serverSocketPath)
	defer conn.Close()

	c, err := client.WriteConfig(ctx, &pb.WriteConfigRequest{Repository: &pb.Repository{StorageName: "default", RelativePath: "foobar.git"}, FullPath: "foo/bar"})
	require.NoError(t, err)
	require.NotNil(t, c)
	require.NotEmpty(t, c.GetError())
}
