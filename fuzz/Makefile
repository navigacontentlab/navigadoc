platform_id != uname -s
open != if [ $(platform_id) = Darwin ]; then \
	echo open; else echo xdg-open; fi
pwd != pwd
package != go list ..
package_name != go list -f '{{.Name}}' ..

docker_run=docker run \
	--rm \
	-v $(pwd)/data:/data \
	-v $(shell dirname $(pwd)):/go/src/$(package) \
	-v /tmp/:/.cache \
	-w /go/src/$(package) \
	$(package_name)-fuzz

.PHONY: clean
clean:
	rm -fr .build-image fuzz/binary.zip

../vendor: ../go.mod ../go.sum
	go mod vendor

.build-image: ../vendor Dockerfile
	cd .. && docker build -t $(package_name)-fuzz -f fuzz/Dockerfile .
	docker inspect --format='{{.Id}}' $(package_name)-fuzz:latest > .build-image

binary.zip: .build-image
	$(docker_run) go-fuzz-build -o fuzz/binary.zip

.PHONY: fuzz
fuzz: .build-image binary.zip
	$(docker_run) go-fuzz -dumpcover \
		-bin=fuzz/binary.zip \
		-workdir=/data

.PHONY: fuzz-coverage
fuzz-coverage: data/cover.html
	$(open) data/cover.html

data/cover.html: data/coverprofile
	$(docker_run) make -C fuzz _fuzz-coverage

.PHONY: _fuzz-coverage
_fuzz-coverage:
	sed -i '/0.0,1.1/d' /data/coverprofile
	go tool cover -html=/data/coverprofile -o /data/cover.html
