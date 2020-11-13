# StackRelease

The `StackRelease` resource is built and managed by the StackApps controller.  
It is namespaced and contains the `AppRevision` that its controller will 
deploy and the relevant configuration taken from the `StackAppConfig`.

```yaml
apiVersion: features.criticalstack.com/v1alpha1
kind: StackRelease
metadata:
  name: demoapp-v1
spec:
  appname: demoapp-v1
  apprevision: # nested AppRevision, see AppRevision docs.
  releaseconfig: # nested config for StackRelease. See StackAppConfig docs.
```
