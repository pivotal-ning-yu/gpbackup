all: depend build test

SHELL := /bin/bash
.DEFAULT_GOAL := all
BACKUP=gpbackup
RESTORE=gprestore
HELPER=gpbackup_helper
DIR_PATH=$(shell dirname `pwd`)
BIN_DIR=$(shell echo $${GOPATH:-~/go} | awk -F':' '{ print $$1 "/bin"}')

GIT_VERSION := $(shell git describe --tags | perl -pe 's/(.*)-([0-9]*)-(g[0-9a-f]*)/\1+dev.\2.\3/')
BACKUP_VERSION_STR="-X github.com/greenplum-db/gpbackup/backup.version=$(GIT_VERSION)"
RESTORE_VERSION_STR="-X github.com/greenplum-db/gpbackup/restore.version=$(GIT_VERSION)"
HELPER_VERSION_STR="-X github.com/greenplum-db/gpbackup/helper.version=$(GIT_VERSION)"
# note that /testutils is not a production directory, but has unit tests to validate testing tools
SUBDIRS_HAS_UNIT=backup/ backup_filepath/ backup_history/ helper/ options/ restore/ utils/ testutils/
SUBDIRS_ALL=$(SUBDIRS_HAS_UNIT) integration/ end_to_end/

DEST = .

GOFLAGS :=

CUSTOM_BACKUP_DIR ?= "/tmp"

.PHONY : coverage integration end_to_end

dependencies :
		go get github.com/alecthomas/gometalinter
		gometalinter --install
		curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
		dep ensure -v
		@cd vendor/golang.org/x/tools/cmd/goimports; go install .
		@cd vendor/github.com/onsi/ginkgo/ginkgo; go install .

format :
		goimports -w .

lint :
		! goimports -l $(SUBDIRS_ALL) | read
		gometalinter --config=gometalinter.config -s vendor ./...  --deadline=120s

unit :
		ginkgo -r -keepGoing -randomizeSuites -noisySkippings=false -randomizeAllSpecs $(SUBDIRS_HAS_UNIT) 2>&1

integration :
		ginkgo -r -randomizeSuites -noisySkippings=false -randomizeAllSpecs integration 2>&1

test : lint unit integration

end_to_end : build
		ginkgo -r -randomizeSuites -slowSpecThreshold=10 -noisySkippings=false -randomizeAllSpecs end_to_end -- --custom_backup_dir $(CUSTOM_BACKUP_DIR) 2>&1

coverage :
		@./show_coverage.sh

depend : dependencies

build :
		CGO_ENABLED=0 go build -tags '$(BACKUP)' $(GOFLAGS) -o $(BIN_DIR)/$(BACKUP) -ldflags $(BACKUP_VERSION_STR)
		CGO_ENABLED=0 go build -tags '$(RESTORE)' $(GOFLAGS) -o $(BIN_DIR)/$(RESTORE) -ldflags $(RESTORE_VERSION_STR)
		CGO_ENABLED=0 go build -tags '$(HELPER)' $(GOFLAGS) -o $(BIN_DIR)/$(HELPER) -ldflags $(HELPER_VERSION_STR)
		@$(MAKE) install_helper helper_path=$(BIN_DIR)/$(HELPER)

build_linux :
		env GOOS=linux GOARCH=amd64 go build -tags '$(BACKUP)' $(GOFLAGS) -o $(BACKUP) -ldflags $(BACKUP_VERSION_STR)
		env GOOS=linux GOARCH=amd64 go build -tags '$(RESTORE)' $(GOFLAGS) -o $(RESTORE) -ldflags $(RESTORE_VERSION_STR)
		env GOOS=linux GOARCH=amd64 go build -tags '$(HELPER)' $(GOFLAGS) -o $(HELPER) -ldflags $(HELPER_VERSION_STR)

build_mac :
		env GOOS=darwin GOARCH=amd64 go build -tags '$(BACKUP)' $(GOFLAGS) -o $(BACKUP) -ldflags $(BACKUP_VERSION_STR)
		env GOOS=darwin GOARCH=amd64 go build -tags '$(RESTORE)' $(GOFLAGS) -o $(RESTORE) -ldflags $(RESTORE_VERSION_STR)
		env GOOS=darwin GOARCH=amd64 go build -tags '$(HELPER)' $(GOFLAGS) -o $(HELPER) -ldflags $(HELPER_VERSION_STR)

install_helper :
		@psql -t -d template1 -c 'select distinct hostname from gp_segment_configuration where content != -1' > /tmp/seg_hosts 2>/dev/null; \
		if [ $$? -eq 0 ]; then \
			gpscp -f /tmp/seg_hosts $(helper_path) =:$(GPHOME)/bin/$(HELPER); \
			if [ $$? -eq 0 ]; then \
				echo 'Successfully copied gpbackup_helper to $(GPHOME) on all segments'; \
			else \
				echo 'Failed to copy gpbackup_helper to $(GPHOME)'; \
				exit 1;	 \
			fi; \
		else \
			echo 'Database is not running, please start the database and run this make target again'; \
				exit 1;	 \
		fi; \
		rm /tmp/seg_hosts

clean :
		# Build artifacts
		rm -f $(BIN_DIR)/$(BACKUP) $(BACKUP)
		rm -f $(BIN_DIR)/$(RESTORE) $(RESTORE)
		rm -f $(BIN_DIR)/$(HELPER) $(HELPER)
		# Test artifacts
		rm -rf /tmp/go-build*
		rm -rf /tmp/gexec_artifacts*
		rm -rf /tmp/ginkgo*
		# Code coverage files
		rm -rf /tmp/cover*
		rm -rf /tmp/unit*
