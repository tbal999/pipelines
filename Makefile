GO_VERSION=1.21.5
GO_TARBALL=go$(GO_VERSION).linux-amd64.tar.gz
GO_URL=https://go.dev/dl/$(GO_TARBALL)
INSTALL_DIR=/usr/local
PROFILE_FILE=$(HOME)/.profile

.PHONY: all install-go update-profile clean

all: install-go update-profile

install-go:
	@if [ -d "$(INSTALL_DIR)/go" ]; then \
		echo "Go is already installed in $(INSTALL_DIR)/go. Skipping download."; \
	else \
		echo "Downloading Go $(GO_VERSION)..."; \
		wget -q $(GO_URL) -O /tmp/$(GO_TARBALL); \
		sudo rm -rf $(INSTALL_DIR)/go; \
		sudo tar -C $(INSTALL_DIR) -xzf /tmp/$(GO_TARBALL); \
		rm /tmp/$(GO_TARBALL); \
		echo "Go installed successfully."; \
	fi

update-profile:
	@if grep -q 'export PATH=.*go/bin' $(PROFILE_FILE); then \
		echo "Go paths already set in $(PROFILE_FILE)."; \
	else \
		echo "Updating $(PROFILE_FILE) with Go environment variables..."; \
		echo '\n# Go environment' >> $(PROFILE_FILE); \
		echo 'export PATH=$$PATH:/usr/local/go/bin' >> $(PROFILE_FILE); \
		echo 'export GOPATH=$$HOME/go' >> $(PROFILE_FILE); \
		echo 'export PATH=$$PATH:$$GOPATH/bin' >> $(PROFILE_FILE); \
		echo "Done. Run 'source $(PROFILE_FILE)' to update your environment."; \
	fi

clean:
	sudo rm -rf /usr/local/go
	rm -f /tmp/$(GO_TARBALL)
	sed -i '/# Go environment/,+3d' $(PROFILE_FILE)
