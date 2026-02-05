#   Copyright Mycophonic.

#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#   You may obtain a copy of the License at

#       http://www.apache.org/licenses/LICENSE-2.0

#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.

NAME := saprobe-dsd
ALLOWED_LICENSES := Apache-2.0,BSD-2-Clause,BSD-3-Clause,MIT
LICENSE_IGNORES := --ignore gotest.tools

include hack/common.mk

# Tests live in a separate module (tests/go.mod) to keep agar out of the
# library go.mod. Override common.mk test targets to include both modules.
TEST_MODULE := $(PROJECT_DIR)/tests

test-unit:
	$(call title, $@)
	@go test $(VERBOSE_FLAG) -count 1 $(PROJECT_DIR)/...
	@cd $(TEST_MODULE) && go test $(VERBOSE_FLAG) -count 1 ./...
	$(call footer, $@)

test-unit-bench:
	$(call title, $@)
	@go test $(VERBOSE_FLAG) -count 1 $(PROJECT_DIR)/... -bench=.
	@cd $(TEST_MODULE) && go test $(VERBOSE_FLAG) -count 1 ./... -bench=.
	$(call footer, $@)

test-unit-race:
	$(call title, $@)
	@CGO_ENABLED=1 go test $(VERBOSE_FLAG) -ldflags="-linkmode=external" -timeout 30m $(PROJECT_DIR)/... -race
	@cd $(TEST_MODULE) && CGO_ENABLED=1 go test $(VERBOSE_FLAG) -ldflags="-linkmode=external" -timeout 30m ./... -race
	$(call footer, $@)

lint-mod:
	$(call title, $@)
	@cd $(PROJECT_DIR) && go mod tidy --diff
	@cd $(TEST_MODULE) && go mod tidy --diff
	$(call footer, $@)

fix-mod:
	$(call title, $@)
	@cd $(PROJECT_DIR) && go mod tidy
	@cd $(TEST_MODULE) && go mod tidy
	$(call footer, $@)
