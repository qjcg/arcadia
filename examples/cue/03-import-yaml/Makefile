.PHONY: all
all: manifest.yml import

# Download podinfo k8s resources.
manifest.yml: version := 6.1.8
manifest.yml:
	kubectl kustomize https://github.com/stefanprodan/podinfo/kustomize?ref=$(version) > $@

.PHONY: import
import:
	cue import manifest.yml \
	--package 'podinfo' \
	--files \
	--force \
	--path 'KubernetesResources:' \
	--path 'strings.ToLower(kind)' \
	--path 'strings.ToLower(metadata.name)' \
	--path 'deploymentEnv: common:'

.PHONY: clean
clean:
	rm -rf manifest*.cue
