# Installation

### Clone the repository and build the dependencies

```shell
git clone https://github.com/criticalstack/stackapps.git
cd stackapps

helm dependency update ./chart
```

#### Leverage the included helm chart for installation 

```shell
helm install stackapps ./chart
```

#### Developers can utilize tilt (ensure you have tilt and go installed on your system)

```shell
tilt up
```
