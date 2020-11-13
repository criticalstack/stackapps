# Stackapps Guide


### StackApps

Critical Stack Applications- `StackApps` represent the premier, officially
supported mechanism for deploying a Kubernetes application into a Critical Stack
cluster. The `StackApp` is intended to be the main representation of
a deployment/release of an application, at a particular version, running in the
cluster. StackApps are built of kubernetes native resources and kubernetes
custom defined resources. They can be interacted with via the Critical Stack UI
or the Kubectl command line tool.

`StackApps` provide the user with a means of representing an entire running 
application with a single cryptographically signed artifact. This provides a
developer friendly means for an application to be moved between environments 
in a repeatable, verifiable, and auditable fashion.

Once the StackApp has been deployed all interaction with the application is 
handled via the StackApp, resource ownership insures the state of the 
application matches the state defined in the StackApp at all times. Any change
to the application is achieved by building and modifying the application in a
development environment and packaging a new revision. When this new revision 
is deployed in a QA or Production environment the underlying application is 
updated to match the state defined in the new revision. This can be accomplished
via standard Kubernetes rolling update or our stand alone Canary deployment 
`StackRelease`.

#### Packaging
`StackApps` are build from applications running inside a cluster. Currently the 
premier packaging mechanism is built into the Critical Stack UI; however there
is a cli on the roadmap and a StackApp can be built and signed manually
via yaml StackApp manifest and the exporting of included resource manifests
with kubectl.

