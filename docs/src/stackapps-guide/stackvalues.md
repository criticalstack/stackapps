# StackValues

StackValues are custom resources that replace sensitive data or information 
that will be specific to an environment. This includes all Kubernetes Secrets
and some ConfigMaps. When a `StackApp` is applied that contains one or more 
`StackValue`, the StackValue controller uses the contained metadata as well
as information from the `StackAppConfig` to fetch the value and build the 
corresponding Kubernetes resource. Currently StackValues can retrieve data
from  Artifactory, Hashicorp Vault, and Amazon S3. 
If a `StackApp` is packaged using the Critical Stack UI Secrets will 
automatically be replaced and ConfigMaps will be included or replaced 
based on annotations. If these annotations do not exist the user will 
be prompted to add them.  During the packaging of a StackApp all 
necessary metadata is gathered from annotations on the resource 
[see annotations here](../resources/StackValueSecret.md). 
These annotations on Secrets and any Configmap that will 
be replaced by a StackValue are necessary before any StackApp 
can be packaged.

See the details on the `StackValues` resource [here](../resources/StackValueSecret.md)
