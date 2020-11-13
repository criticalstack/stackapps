# Quick Start
### Getting it running.
While the StackApps controllers can be run on their own, the Critical Stack UI
has been designed with them in mind. In addition the Critical Stack UI is
currently the primary way to package a running application into a `StackApp`. 

If you would like to add the UI to your cluster please follow the steps in the UI repository
[here](https://github.com/criticalstack/ui).

In the case of a local cluster or remote development cluster, the Critical
Stack UI and `StackApps` controllers can be used together. For a production
cluster in which applications will be deployed via `StackApps` the Critical
Stack UI is not necessary.

To Install the `StackApps` Controller please follow the
[Installation](./installation.md) page.

For information on creating, packaging, and signing a `StackApp` from a running
application see the docs [here](https://github.com/criticalstack/ui) 


### Deploying an application to a new cluster with StackApps. 
Once you have a packaged and exported `StackApp` you'll need to set up a few things as an
administrator in the cluster in which you wish to deploy the `StackApp`. 

1. Ensure the `StackApps` controllers and resource definitions are on the target
   cluster. The Controllers should be running in the `critical-stack`
   namespace.
2. Apply your `StackAppsConfig` for the application. Details can be found
   [here](./configuration.md).
3. Export the `VerificationKey` from the key pair that was used to sign the
   `StackApp` from the cluster in which it was created. Apply it to the target
   cluster in the namespace specified in the `StackAppsConfig` as
   `appnamespace`.
4. Apply the `StackApp`. The `StackApp` is a cluster scoped resource, you will not
   need to specify a namespace. Once the `StackApp` has been applied allow time
   for all of the resources to become healthy and check out your application
   running in its new home!

