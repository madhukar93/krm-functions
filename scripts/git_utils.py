import os
from git import Repo
from github import Github

GIT_URL = "https://github.com/bukukasio/tokko-k8s"
GIT_REPO="bukukasio/tokko-k8s"
OWNER="bukukasio"
REPO="tokko-k8s"
BASE_BRANCH="master"
AUTHOR="platform@lummo.com"
GIT_USER="lummo-robot"
GIT_TOKEN=os.getenv("GIT_TOKEN")

def git_clone_checkout(branch_name):
    git_url = f"https://{GIT_USER}:{GIT_TOKEN}@github.com/bukukasio/tokko-k8s"
    repo = Repo.clone_from(git_url, f"./{REPO}")
    repo.git.checkout('-b', branch_name)
    return repo

def git_push(repo, branch_name, function_name, new_tag):
    repo.git.add(update=True)
    repo.git.commit('-m', f'krm function version upgrade: Updated the version for {function_name} function with version {new_tag}', author=f'{GIT_USER} <{AUTHOR}>')
    repo.git.push('origin', branch_name)

def create_pull_request(branch_name, function_name, new_tag):
    g = Github(GIT_USER, GIT_TOKEN)
    repo = g.get_repo(GIT_REPO)
    pr = repo.create_pull(
                            title=f"krm function version upgrade for {function_name}",
                            body=f"krm function version upgrade: Updated the version for {function_name} with version {new_tag}",
                            head=f"{branch_name}",
                            base=f"{BASE_BRANCH}"
                        )
    print(pr)
