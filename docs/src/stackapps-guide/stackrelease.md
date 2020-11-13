# StackRelease

The StackRelease is a Namespace scoped custom resource that contains information about
the deployment details for an application. If the StackRelease feature is not 
used the StackRelease is largely a pass through step on and only serves to 
deploy the AppRevision into its namespace. If the StackRelease feature is 
toggled on the StackRelease controller will deploy the application into 
a newly built namespace that is named with the revision and then build the 
resources necessary to preform a canary deployment and slowly shift traffic
into the new namespace from the stable version. The mechanism for this is 
defined in the StackReleaseConfig( ### link doc), Currently Treafik is the 
only supported mechanism however service mesh integration is on the roadmap.

The StackRelease will be deployed and managed by the StackApp
controller; however relevant diagnostic and monitoring information can
be had with introspection. `kubectl describe stackRelease <myStackRelease>`


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

see additional details about the `StackRelease` [here](../resources/StackReleaseResource.md) 
