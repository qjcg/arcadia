go_pkg_version := v0.24.3

# NOTE: Must also get apimachinery package or `cue get go` exits with an error.
# See https://github.com/cue-lang/cue/discussions/1497#discussioncomment-2071985
go_pkgs := k8s.io/api@$(go_pkg_version) k8s.io/apimachinery@$(go_pkg_version)
cue_pkgs := k8s.io/api/core/v1 k8s.io/api/extensions/v1beta1

.PHONY: godefs
godefs:
	cue mod init example.com/mypkg
	go get $(go_pkgs)
	cue get go $(cue_pkgs)

.PHONY: clean
clean:
	rm -rf cue.mod
