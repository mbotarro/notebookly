# notebookly
Datadog notebooks tool

## Usage

```shell

Usage: notebookly <command>

Flags:
  -h, --help     Show context-sensitive help.
      --debug    Enable debug mode.

Commands:
  clone
    Clone a notebook.

./notebookly clone --id $ID --var kube_cluster_name,kube_cluster_name,canteven --var "host,host,*" --start 2022-08-09T18:00:00+07:00 --end 2022-08-09T20:00:00+07:00
```
