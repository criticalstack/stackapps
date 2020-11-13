# -*- mode: Python -*-

load('ext://restart_process', 'docker_build_with_restart')

allow_k8s_contexts('kubernetes-admin@ui-dev')

k8s_yaml(listdir('./chart/crds'))

apiFiles = [f for f in listdir('./api', recursive=True) if not '/zz_generated.' in f]

local_resource('stackapps-objects', 'make generate', deps=[
    'Makefile',
] + apiFiles)
local_resource('stackapps-crds', 'make manifests', resource_deps=['stackapps-objects'], deps=[
    'Makefile',
] + apiFiles)
local_resource('stackapps-manager', 'make build', resource_deps=['stackapps-objects'], deps=[
    'Makefile', 'api', 'go.mod', 'go.sum', 'main.go', 'pkg', 'controllers'
])
local_resource('stackapps-manager-test', 'make test', auto_init=False, trigger_mode=TRIGGER_MODE_MANUAL)

docker_build_with_restart(
    'cscr.io/criticalstack/stackapps',
    '.',
    dockerfile='hack/Dockerfile',
    entrypoint="/manager",
    build_args={'GOPROXY':os.environ.get('GOPROXY', '')},
    only=[
        './scripts/manifests',
        './bin/manager'
    ],
    live_update=[
        sync('./scripts/manifests', '/manifests'),
        sync('./bin/manager', '/manager'),
    ]
)

k8s_yaml(helm(
  './chart',
  # The release name, equivalent to helm --name
  name='cs-stackapps',
  # The namespace to install in, equivalent to helm --namespace
  namespace='critical-stack',
))
