APP_NAME         := vircam
OUT_DIR          := dist
LINUX_OS		 := linux
GOFLAGS_COMMON   := -trimpath -buildvcs=false -mod=readonly
LDFLAGS_COMMON   := -s -w
TAGS_COMMON		 := linux

build-linux-%:
	@echo "==> Building $(APP_NAME) for $(LINUX_OS)/$*"
	@mkdir -p $(OUT_DIR)
	@GOOS=$(LINUX_OS) GOARCH=$* CGO_ENABLED=1 CC=x86_64-linux-gnu-gcc \
		go build $(GOFLAGS_COMMON) \
			-tags $(TAGS_COMMON)  \
			-ldflags '$(LDFLAGS_COMMON)' \
			-o $(OUT_DIR)/$(APP_NAME)-$(LINUX_OS)-$* \
			main.go
	@ls -lh $(OUT_DIR)/$(APP_NAME)-$(LINUX_OS)-$*


.PHONY: build
build: build-linux-amd64