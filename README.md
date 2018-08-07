# kubesecrets

 Manage kubernetes secrets.

# Installation
You find releases in [Releases](releases) section.

## Usage
```
usage: kubesecrets [<flags>] <command> [<args> ...]

Flags:
  --help                 Show context-sensitive help (also try --help-long and --help-man).
  --incluster            use in-cluster authentication
  --namespace="default"  namespace to use
  --output=json          output format to use
  --kubeconfig="/home/tobias/.kube/config"
                         absolute path to the kubeconfig file
  --version              Show application version.

Commands:
  help [<command>...]
    Show help.

  get [<filter>...]
    get a secret

  set [<flags>] <name> <value>
    set a secret

  delete <name>
    delete a secret
```

## Examples
### Create a secret
[//]: # (kubesecrets set api.key "Hello World")
[![create.svg](create.svg)](create.svg)

### Set a secret
[//]: # (kubesecrets set api.token "A Token")
[![set.svg](set.svg)](set.svg)

### Delete a secret key
[//]: # (kubesecrets delete api.key)
[![delete_key.svg](delete_key.svg)](delete_key.svg)


### Delete a secret
[//]: # (kubesecrets delete api)
[![delete.svg](delete.svg)](delete.svg)
