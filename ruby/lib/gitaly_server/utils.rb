module GitalyServer
  module Utils
    def gitaly_commit_from_rugged(rugged_commit)
      message_split = rugged_commit.message.split("\n", 2)
      Gitaly::GitCommit.new(
        id: rugged_commit.oid,
        subject: message_split[0] ? message_split[0].chomp.b : "",
        body: rugged_commit.message.b,
        parent_ids: rugged_commit.parent_ids,
        author: gitaly_commit_author_from_rugged(rugged_commit.author),
        committer: gitaly_commit_author_from_rugged(rugged_commit.committer)
      )
    end

    def gitaly_commit_author_from_rugged(rugged_author)
      Gitaly::CommitAuthor.new(
        name: rugged_author[:name].b,
        email: rugged_author[:email].b,
        date: Google::Protobuf::Timestamp.new(seconds: rugged_author[:time].to_i)
      )
    end

    def commit_details_from_gitaly(gitaly_commit_details)
      Gitlab::Git::Wiki::CommitDetails.new(
        gitaly_commit_details.name,
        gitaly_commit_details.email,
        gitaly_commit_details.message
      )
    end

    def bridge_exceptions
      yield
    rescue GRPC::BadStatus => e
      # Pass GRPC back without wrapping
      raise e
    rescue StandardError => e
      raise GRPC::Unknown.new(e.message, { "gitaly-ruby.exception.class": e.class.name })
    end
  end
end
