module GitalyServer
  class RefService < Gitaly::RefService::Service
    include Utils

    def create_branch(request, _call)
      start_point = request.start_point
      start_point = 'HEAD' if start_point.empty?
      branch_name = request.name

      repo = Gitlab::Git::Repository.from_call(_call)
      rugged_ref = repo.rugged.branches.create(branch_name, start_point)

      Gitaly::CreateBranchResponse.new(
        status: :OK,
        branch: Gitaly::Branch.new(
          name: bytes!(rugged_ref.name),
          target_commit: gitaly_commit_from_rugged(rugged_ref.target),
        ),
      )
    rescue Rugged::ReferenceError => e
      status = case e.to_s
               when /'refs\/heads\/#{branch_name}' is not valid/
                 :ERR_INVALID
               when /a reference with that name already exists/
                 :ERR_EXISTS
               else
                 :ERR_INVALID_START_POINT
               end

      Gitaly::CreateBranchResponse.new(status: status)
    end

    def delete_branch(request, _call)
      branch_name = request.name
      raise GRPC::InvalidArgument.new("empty Name") if branch_name.empty?

      repo = Gitlab::Git::Repository.from_call(_call)
      repo.delete_branch(branch_name)

      Gitaly::DeleteBranchResponse.new
    rescue Rugged::ReferenceError => e
      raise GRPC::Internal.new(e.to_s)
    end
  end
end
