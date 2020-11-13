# stackapps helm chart

Installs CRDs and controller(s) for StackApps.

If `webhooks.enabled=true` (default is `false`), validating and defaulting webhooks will be installed - this relies on [cert-manager](https://cert-manager.io/docs/installation/kubernetes/) being installed in your cluster.
