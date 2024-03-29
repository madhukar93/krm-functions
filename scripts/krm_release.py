import argparse
from file_utils import modify_files
from git_utils import git_clone_checkout, git_push, create_pull_request

def main():
    arg_parser = argparse.ArgumentParser(description='Get the function name as input and update new release in the files accordingly')
    arg_parser.add_argument('function',
                        metavar='function',
                        type=str,
                        help='function name to update the new version tag')
    arg_parser.add_argument('new_tag',
                        metavar='new_version',
                        type=str,
                        help='new version tag to be updated in the files')

    args = arg_parser.parse_args()

    function_name = args.function
    new_tag = args.new_tag
    branch_name=f"KRM-Func-upgrade-{function_name}-{new_tag}"

    repo = git_clone_checkout(branch_name)
    modify_files(function_name=function_name, new_tag=new_tag)
    git_push(repo, branch_name, function_name, new_tag)
    create_pull_request(branch_name, function_name, new_tag)

if __name__ == "__main__":
    main()
