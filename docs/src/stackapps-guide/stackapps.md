# StackApps

The StackApp is the Highest order resource in the StackApps ecosystem, and the
only resource a developer (or ci pipeline) needs to interact with. StackApps are
immutable and changes should be represented in new revisions. 

When a StackApp is bundled it contains two major components. The meat 
of the StackApp is in the nested AppRevision, the other main component that 
is a configmap containing a grouping of all of the manifests that 
make up the application in the form of a configmap. There are additional 
elements that will be populated on deployment by the StackApps controller.

The manifests are signed on StackApp creation and result is stored in the 
AppRevision. StackApps are cluster scoped resources, for additional information 
on AppRevisions see the AppRevision Documentation. 

See the details on the `StackApp` resource [here](../resources/stackAppsResource.md)
