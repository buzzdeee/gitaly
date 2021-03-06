package notifications

import pb "gitlab.com/gitlab-org/gitaly-proto/go"

type server struct{}

// NewServer creates a new instance of a grpc NotificationsServer
func NewServer() pb.NotificationServiceServer {
	return &server{}
}
