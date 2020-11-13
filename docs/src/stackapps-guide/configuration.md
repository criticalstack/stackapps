# Configuration

Configuration of StackApps is handled through a cluster scoped custom resource of *KIND* `StackAppConfig` in the *API-GROUP* `features.criticalstack.com` and is application specific.

the Spec of the `StackAppConfig` contains basic configuration for the application and more specific configuration for each StackApps controller.

```yaml
apiVersion: features.criticalstack.com/v1alpha1
kind: StackAppConfig
metadata:
  generation: 1
  name: stackapp-config
spec:
  environmentname: dev
  appnamespace: prod-app
  stackapps: #StackApps specific configuration
  stackvalues: #StackValues Specific Configuration
  releases: #StackRelease Specific Configuration 
```

### StackApps Specific Configuration
```yaml
  stackapps:
    enabled: false
```
This is a switch to disable the StackApps Controllers and currently is not used


### StackValues Specific Configuration
```yaml
  stackvalues:
    enabled: false
    insecure: false
    tokenName: vault-token
    sources: #array of source configurations
```

`enabled`: Currently not implemented.

`insecure`: uses the hardcoded values provided in annotation form to populate the k8s resources. This is only for testing with non-secure text in a development invironment.

`tokenName`: the name of the k8s secret that contains the tokens to authenticate to platforms in which secrets and configuration items will be stored.

`sources`: Array of source configurations. Currently supported source types include `artifactory`, `aws_s3`, and `vault`.


#### StackValue Sources
```yaml
      vault:
        name: myvault
        region: west
        type: vault
        route: "http://vault.external.svc:8200"
```

`name`: alais to be used for the source

'region`: Aws hosting region, only used if source is equal to `aws_s3`

`type`: type of source 

`route`: url in which source is hosted 

### StackRelease Configuration  

```yaml
    enabled: true
    backendType: traefik
    ingressPort: 30080
    host: myapp.com
    releaseStages:
    - canaryWeight: 20
      stepDuration: "1m"
    - canaryWeight: 50
      stepDuration: "1m"
    - canaryWeight: 100 
```

`enabled`: Currently not implemented.

`backendType`: Specifies infrastructrue to use for deployends. Currently `traefik` is the only supported value. if this field does not exist StackRelease will not be used.

`ingressPort`: Port in which the application will be exposed.

`host`: Hostname of the application.

`releaseStages`: Array of stages the canary controller will step though. `canaryWeight` is represented in percent and `stepDuration is represented as a time i.e. "60s" is equal to "1m".


this full example yaml file can be found this link that i might put here at some point 
