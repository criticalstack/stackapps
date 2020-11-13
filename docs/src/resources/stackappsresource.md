# StackApp

```yaml
apiVersion: features.criticalstack.com/v1alpha1
kind: StackApp
metadata:
  generation: 1
  name: demoapp-v1
spec:
  appRevision: # nested appRevision See appRevision docs.
  majorVersion: 1
```

#### StackApp Spec
`appRevision`: contains details about the application being deployed, see 
AppRevision Docs.

`majorVersion`: Major version of the application. This is Necessary because Major
Version is using in naming to insure multiple major versions can be run 
independently 


#### StackApp ConfigMap
```yaml
apiVersion: v1
data:
  manifests: | # note pipe to allow multiline string
kind: ConfigMap
metadata:
  labels:
    stackapps.criticalstack.com/export: ignore
  name: demoapp-v1-r1
```

`manifests`: list of yaml kubernetes resources separated by the standard `---`.  
