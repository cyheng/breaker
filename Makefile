export GO111MODULE=on
LDFLAGS := -s -w

os-archs=linux:amd64 windows:amd64

all: build

build:  bridge portal

bridge:
	go build -o bin/bridge ./cmd/bridge

portal:
	go build -o bin/portal ./cmd/portal

linux:
	@$(foreach n, $(os-archs),\
		os=$(shell echo "$(n)" | cut -d : -f 1);\
		arch=$(shell echo "$(n)" | cut -d : -f 2);\
		target_suffix=$${os}_$${arch};\
		echo "Build $${os}-$${arch}...";\
		env CGO_ENABLED=0 GOOS=$${os} GOARCH=$${arch}  go build -trimpath -ldflags "$(LDFLAGS)" -o ./release/bridge_$${target_suffix} ./cmd/bridge;\
		env CGO_ENABLED=0 GOOS=$${os} GOARCH=$${arch}  go build -trimpath -ldflags "$(LDFLAGS)" -o ./release/portal_$${target_suffix} ./cmd/portal;\
		echo "Build $${os}-$${arch} done";\
	)