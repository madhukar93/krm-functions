from git import Repo

GIT_URL = "https://github.com/bukukasio/tokko-k8s"

def git_clone():
    Repo.clone_from(GIT_URL, "./tokko-k8s")

def git_commit():
    pass

def git_push():
    pass

def git_pr():
    pass
