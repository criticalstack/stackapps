# AppRevision



```yaml
apiVersion: features.criticalstack.com/v1alpha1
kind: AppRevision
metadata:
  name: demoapp-v1
spec:
  appRevisionConfig: #nested from StackAppConfig
  manifests: demoapp-v1-r1
  revision: 1
  signatures: #RSA signature(s)
```

