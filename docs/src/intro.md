
StackApps Overview
------------------

A `StackApp` is a resource in the Critical Stack cluster which repesents
a kubernetes *application* at a particular version.

The `StackApp` contains:

  - Metadata about the *application*
  - A reference to the *manifests* used to deploy all associated resources
  - Signature details for the *manifests*

the Critical Stack cluster contains infrastructure that is aware of *StackApps*.
When the `StackApp` resource is created or modified, the underlying
*application* resources are updated accordingly (assuming that the cluster's
configured deployment guarantees are met which may include signing,
compatibility checks, etc).

A `StackApp` may be created from raw kubernetes *manifests* that already exist
(or that are generated from existing source material), or it may be created by
encoding selected objects running within a kubernetes cluster namespace (see
[packaging](./tech-details.md#packaging)). The former method (existing
*manifests*) is intended to serve as a bridge for developers with existing
kubernetes *applications*, or those who would like to create their applications
using third-party tools. The latter (running resources) is tailored toward
developers creating their *applications* within the Critical Stack "ecosystem"
to begin with - e.g. with the aid of CS UI and accompanying tools. The
Critical Stack API and UI would provide functionality for selecting running resources
for packaging and export into a `StackApp`.

Once a `StackApp` has been deployed (or created from components running in the
cluster), the CS UI can provide aggregated status information (health checks,
metrics data, etc) for `StackApps` in each namespace. Because this system is
fully integrated with the kubernetes API, this information is easily available
to other tools such as `kubectl` without the need for the CS UI to act as an
intermediary.

Resources that belong to a `StackApp` are protected (using kubernetes RBAC
permissions) by default. Similarly, all *manifests* currently or previously
referenced by a `StackApp` are immutable.


![Overview diagram](./stackapps-overview.png)
