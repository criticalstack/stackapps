# StackValue


### StackValue Annotations

`stackvalues.criticalstack.com/stackValue`: if true, build StackValue from
ConfigMap and omit it from the stackapp.

`stackvalues.criticalstack.com/sourceType`: Type of source value will be retrieved 
from. Should be one of the supported types (Artifactory, Vault, or AWS_S3).

`stackvalues.criticalstack.com/path`: Endpoint required to retrieve value. The 
base URL is defined in the `StackAppsConfig`. 

`stackvalues.criticalstack.com/insecureval`: non-Secure value to be used if the 
StackApp is deployed to a development Cluster.

### Example Secret prepared for StackValues

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: demoapp-db-credentials
  annotations:
    stackvalues.criticalstack.com/path: "v1/secret/data/myapp/password"
    stackvalues.criticalstack.com/sourceType: "vault"
    stackvalues.criticalstack.com/insecureval: "password"
data:
  value: MWYyZDFlMmU2N2Rm
```

### Resulting StackValue that will be included in the StackApp
```yaml
kind: StackValue
metadata:
  name: demoapp-db-credentials
spec:
  insecureVal: password
  name: demoapp-db-credentials
  objectType: Secret
  path: v1/secret/data/myapp/password
  sourceType: vault
```


When this StackValue is applied to the cluster the StackValue controller will 
reconcile it into a kubernetes Secret. The Value will be retrieved by 
an api call to Vault at the URL provided for Vault in the StackAppsConfig
at the api endpoint defined in `path:` above.

Note that this is handled this way because the CI pipeline or developer
that apply the StackApp should not have the ability to define an external
location for making API calls. Access to the `StackAppsConfig` should be 
limited to administrators via RBAC.

