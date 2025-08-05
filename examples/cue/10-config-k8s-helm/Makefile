out := grafana*

.PHONY: all
all: charts imports

.PHONY: charts
charts:
	helm repo add bitnami https://charts.bitnami.com/bitnami
	helm pull --untar bitnami/grafana

.PHONY: imports
imports:
	cue import grafana/values.yaml --package grafana --path '#CommonValues:' --force --outfile grafana-values.common.cue
	cue import grafana/values.yaml --package grafana --path '#DevValues:' --force --outfile grafana-values.dev.cue
	cue import grafana/values.yaml --package grafana --path '#UATValues:' --force --outfile grafana-values.uat.cue
	cue import grafana/values.yaml --package grafana --path '#ProdValues:' --force --outfile grafana-values.prod.cue

.PHONY: cluster
cluster:
	k3d cluster create --servers 3 --agents 3 --port 8888:3000@loadbalancer helm-example


.PHONY: cluster-install
cluster-install:
	helm install --create-namespace --namespace dev -f grafana-values.common.yaml grafana-dev bitnami/grafana
	helm install --create-namespace --namespace uat -f grafana-values.common.yaml grafana-uat bitnami/grafana
	helm install --create-namespace --namespace prod -f grafana-values.common.yaml grafana-prod bitnami/grafana

.PHONY: cluster-delete
cluster-delete:
	k3d cluster delete helm-example

.PHONY: clean
clean:
	rm -rf $(out)
