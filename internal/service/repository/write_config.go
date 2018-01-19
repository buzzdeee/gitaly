package repository

import (
	"golang.org/x/net/context"

	pb "gitlab.com/gitlab-org/gitaly-proto/go"
	"gitlab.com/gitlab-org/gitaly/internal/rubyserver"
)

func (s *server) WriteConfig(ctx context.Context, req *pb.WriteConfigRequest) (*pb.WriteConfigResponse, error) {
	client, err := s.RepositoryServiceClient(ctx)
	if err != nil {
		return nil, err
	}

	// We handle the "no repository" error separately inside ruby...
	clientCtx, err := rubyserver.SetHeadersWithoutRepoCheck(ctx, req.GetRepository())
	if err != nil {
		return nil, err
	}

	return client.WriteConfig(clientCtx, req)
}
