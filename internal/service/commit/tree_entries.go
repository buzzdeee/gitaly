package commit

import (
	"bufio"
	"fmt"
	"io"

	log "github.com/sirupsen/logrus"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	pb "gitlab.com/gitlab-org/gitaly-proto/go"
	"gitlab.com/gitlab-org/gitaly/internal/git/catfile"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var maxTreeEntries = 1000

func validateGetTreeEntriesRequest(in *pb.GetTreeEntriesRequest) error {
	if len(in.GetRevision()) == 0 {
		return fmt.Errorf("empty Revision")
	}

	if len(in.GetPath()) == 0 {
		return fmt.Errorf("empty Path")
	}

	return nil
}

func populateFlatPath(entries []*pb.TreeEntry, stdin io.Writer, stdout *bufio.Reader) error {
	for _, entry := range entries {
		entry.FlatPath = entry.Path

		if entry.Type != pb.TreeEntry_TREE {
			continue
		}

		for {
			subentries, err := treeEntries(entry.CommitOid, string(entry.FlatPath), stdin, stdout)

			if err != nil {
				return err
			}
			if len(subentries) != 1 || subentries[0].Type != pb.TreeEntry_TREE {
				break
			}

			entry.FlatPath = subentries[0].Path
		}
	}

	return nil
}

func getTreeEntriesHandler(stream pb.CommitService_GetTreeEntriesServer, revision, path string) catfile.Handler {
	return func(stdin io.Writer, stdout *bufio.Reader) error {
		entries, err := treeEntries(revision, path, stdin, stdout)
		if err != nil {
			return err
		}

		if err := populateFlatPath(entries, stdin, stdout); err != nil {
			return err
		}

		for len(entries) > maxTreeEntries {
			chunk := &pb.GetTreeEntriesResponse{
				Entries: entries[:maxTreeEntries],
			}
			if err := stream.Send(chunk); err != nil {
				return err
			}
			entries = entries[maxTreeEntries:]
		}

		if len(entries) > 0 {
			return stream.Send(&pb.GetTreeEntriesResponse{Entries: entries})
		}

		return nil
	}
}

func (s *server) GetTreeEntries(in *pb.GetTreeEntriesRequest, stream pb.CommitService_GetTreeEntriesServer) error {
	grpc_logrus.Extract(stream.Context()).WithFields(log.Fields{
		"Revision": in.Revision,
		"Path":     in.Path,
	}).Debug("GetTreeEntries")

	if err := validateGetTreeEntriesRequest(in); err != nil {
		return status.Errorf(codes.InvalidArgument, "TreeEntry: %v", err)
	}

	revision := string(in.GetRevision())
	path := string(in.GetPath())
	handler := getTreeEntriesHandler(stream, revision, path)

	return catfile.CatFile(stream.Context(), in.Repository, handler)
}
