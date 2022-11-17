import requests, json
from git import Repo
from github import Github


GIT_URL = "https://github.com/bukukasio/tokko-k8s"
GIT_REPO="bukukasio/tokko-k8s"
OWNER="bukukasio"
REPO="tokko-k8s"
GIT_USER=""
BASE_BRANCH="master"
AUTHOR=""

def git_clone_checkout(branch_name):
    repo = Repo.clone_from(GIT_URL, f"./{REPO}")
    repo.git.checkout('-b', branch_name)
    return repo

def git_push(repo, branch_name, function_name, new_tag):
    repo.git.add(update=True)
    repo.git.commit('-m', f'krm function version upgrade: Updated the version for {function_name} function with version {new_tag}', author=f'{AUTHOR}')
    repo.git.push('origin', branch_name)

def create_pull_request(git_token, branch_name, function_name, new_tag):
    g = Github(git_token)
    repo = g.get_repo(GIT_REPO)
    pr = repo.create_pull(
                            title=f"krm function version upgrade for {function_name}",
                            body=f"krm function version upgrade: Updated the version for {function_name} with version {new_tag}",
                            head=f"{GIT_USER}:{branch_name}",
                            base=f"{BASE_BRANCH}"
                        )
    print(pr)
