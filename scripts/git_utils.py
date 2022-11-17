import requests, json
from git import Repo

GIT_URL = "https://github.com/bukukasio/tokko-k8s"
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
    try:
        repo.git.add(update=True)
        repo.git.commit('-m', f'krm function version upgrade: Updated the version for {function_name} function with version {new_tag}', author=f'{AUTHOR}')
        repo.git.push('origin', branch_name)
    except Exception as ex:
        print(ex)

def create_pull_request(git_token, branch_name, function_name, new_tag):
    headers = {
        'Accept': 'application/vnd.github+json',
        'Authorization': f'Bearer {git_token}',
        'Content-Type': 'application/x-www-form-urlencoded',
        }
    data = {
            "title":"krm function version upgrade",
            "body": f"krm function version upgrade: Updated the version for {function_name} with version {new_tag}",
            "head": f"{GIT_USER}:{branch_name}",
            "base": f"{BASE_BRANCH}"
        }
    response = requests.post(f'https://api.github.com/repos/{OWNER}/{REPO}/pulls', headers=headers, data=json.dumps(data))
    print(response.text)
    return response
