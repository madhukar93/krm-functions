# example

## Description
this is an example blueprint

## Usage

### Fetch the package
`kpt pkg get REPO_URI[.git]/PKG_PATH[@VERSION] example`
Details: https://kpt.dev/reference/cli/pkg/get/

### View package content
`kpt pkg tree example`
Details: https://kpt.dev/reference/cli/pkg/tree/

### Apply the package
```
kpt live init example
kpt live apply example --reconcile-timeout=2m --output=table
```
Details: https://kpt.dev/reference/cli/live/
