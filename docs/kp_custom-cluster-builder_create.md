## kp custom-cluster-builder create

Create a custom cluster builder

### Synopsis

Create a custom cluster builder by providing command line arguments.
This custom cluster builder will be created only if it does not exist.

Tag when not specified, defaults to a combination of the canonical repository and specified builder name.
The canonical repository is read from the "canonical.repository" key in the "kp-config" ConfigMap within "kpack" namespace.


```
kp custom-cluster-builder create <name> [flags]
```

### Examples

```
kp ccb create my-builder --order /path/to/order.yaml --stack tiny --store my-store
kp ccb create my-builder --order /path/to/order.yaml
kp ccb create my-builder --tag my-registry.com/my-builder-tag --order /path/to/order.yaml --stack tiny --store my-store
kp ccb create my-builder --tag my-registry.com/my-builder-tag --order /path/to/order.yaml
```

### Options

```
  -h, --help           help for create
  -o, --order string   path to buildpack order yaml
  -s, --stack string   stack resource to use (default "default")
      --store string   buildpack store to use (default "default")
  -t, --tag string     registry location where the builder will be created
```

### SEE ALSO

* [kp custom-cluster-builder](kp_custom-cluster-builder.md)	 - Custom Cluster Builder Commands

###### Auto generated by spf13/cobra on 30-Jul-2020
