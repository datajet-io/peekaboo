BUILD_SUFFIX=$(shell basename ${PWD})
DOCKER_IMAGE=datajet/peekaboo
GIT_REV=$(shell git rev-parse --short HEAD)
GOPACKAGES=$(shell glide nv)
GO_VERSION=$(or $(GIMME_GO_VERSION),1.7.1)
OUTPUTS_BASEDIR=$(or $(CIRCLE_ARTIFACTS),/tmp)
OUTPUTS_PATH=$(shell echo "${OUTPUTS_BASEDIR}/${BUILD_SUFFIX}")
DATESTAMP=$(shell date +%Y%m%d)
VERSION=$(shell echo "${DATESTAMP}_${GIT_REV}")

.PHONY: build
build:
	go install -v 2>&1 | tee $(OUTPUTS_PATH)_goinstall.log

.PHONY: build-full
build-full:
	go build -v 2>&1 | tee $(OUTPUTS_PATH)_gobuild.log

.PHONY: docker-build
docker-build:
	docker build -t $(DOCKER_IMAGE):$(VERSION) . | tee $(OUTPUTS_PATH)_dockerbuild.log

.PHONY: docker-push
docker-push:
	docker login -e $(DOCKER_EMAIL) -u $(DOCKER_USER) -p $(DOCKER_PASS)
	docker push $(DOCKER_IMAGE):$(VERSION) . | tee $(OUTPUTS_PATH)_dockerpush.log
	docker push $(DOCKER_IMAGE):latest . | tee -a $(OUTPUTS_PATH)_dockerpush.log

.PHONY: ci
ci: setup lint build-full docker-build

.PHONY: ci-setup
ci-setup:
	scripts/ci-setup-common

.PHONY: gimme
gimme:
	@cat ${HOME}/.gimme/envs/go${GO_VERSION}.env

.PHONY: lint
lint:
	gometalinter $(GOPACKAGES) 2>&1 | tee $(OUTPUTS_PATH)_gometalinter.log || true

.PHONY: setup
setup:
	glide install 2>&1 | tee $${CIRCLE_ARTIFACTS:-/tmp}/$(BUILD_SUFFIX)_glide.log

.PHONY: test
test:
	go test -v $(GOPACKAGES) 2>&1 | tee $(OUTPUTS_PATH)_gotest.log

.PHONY: testsuite
testsuite:
	gocov test -v $(GOPACKAGES) | tee $(OUTPUTS_PATH)_gocov.log | gocov-html > $(OUTPUTS_PATH)_coverage.html
	go vet -v $(GOPACKAGES) 2>&1 | tee $(OUTPUTS_PATH)_govet.log
	gofmt -l `glide nv | sed 's/\.\.\././g'`
