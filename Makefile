.PHONY: setup lint-tools lint format check validate-deps install-prereqs install build validate clean clean-all aspire-run aspire-watch troubleshoot kill-dev help all

# VERBOSITY CONTROL
Q := @
ifdef VERBOSE
  Q :=
endif

# SERVICE DIRECTORIES
APPHOST_DIR := services/apphost
AUTH_DIR    := services/authingway
NAUR_DIR    := services/naurffxiv
MOD_DIR     := services/moddingway
FIND_DIR    := services/findingway
CLEAR_DIR   := services/clearingway

# PROJECT FILES
APPHOST_PRJ := $(APPHOST_DIR)/Naur.AppHost.csproj
AUTH_PRJ    := $(AUTH_DIR)/Naur.Authingway.csproj

# SCRIPTS & COMMANDS
SCRIPTS_DIR := scripts/makefile
PWSH 		:= pwsh -NoProfile -ExecutionPolicy Bypass -File
WIN_PWSH    := powershell -NoProfile -ExecutionPolicy Bypass -File
PYTHON_RAW  := $(firstword $(wildcard services/moddingway/venv/Scripts/python.exe) $(wildcard services/moddingway/venv/bin/python) python3 python)
ifeq ($(OS),Windows_NT)
    PYTHON := $(subst /,\,$(PYTHON_RAW))
else
    PYTHON := $(PYTHON_RAW)
endif

# DEFAULT TARGET
.DEFAULT_GOAL := help

# AGGREGATE TARGETS
all: setup install build

# SETUP & INSTALLATION
setup:
	$(Q)$(WIN_PWSH) $(SCRIPTS_DIR)/setup.ps1
	$(Q)echo [TIP] Restart your terminal if tools were installed.

lint-tools:
	$(Q)$(PWSH) $(SCRIPTS_DIR)/lint-tools.ps1

lint:
	$(Q)$(PWSH) $(SCRIPTS_DIR)/lint.ps1

format:
	$(Q)$(PWSH) $(SCRIPTS_DIR)/lint.ps1 -Fix

check:
	$(Q)$(PWSH) $(SCRIPTS_DIR)/check.ps1

# OS DETECTION
ifeq ($(OS),Windows_NT)
    NULL_DEVICE := NUL
else
    NULL_DEVICE := /dev/null
endif

check-python:
	$(Q)$(PYTHON) --version >$(NULL_DEVICE) 2>&1 || (echo "Python required." && exit 1)

validate-deps: check-python
	$(Q)echo Validating that dependencies.yml matches manifests...
	$(Q)$(PYTHON) scripts/check-dependency-sync.py



install: check lint-tools
	$(Q)$(PWSH) $(SCRIPTS_DIR)/install.ps1 \
		-AppHostPrj $(APPHOST_PRJ) \
		-AuthPrj $(AUTH_PRJ) \
		-NaurDir $(NAUR_DIR) \
		-ModDir $(MOD_DIR) \
		-FindDir $(FIND_DIR) \
		-ClearDir $(CLEAR_DIR)
	$(Q)$(MAKE) validate-deps

# BUILD
build:
	$(Q)$(PWSH) $(SCRIPTS_DIR)/build-all.ps1 \
		-AppHostPrj $(APPHOST_PRJ) \
		-AuthPrj $(AUTH_PRJ) \
		-NaurDir $(NAUR_DIR) \
		-ModDir $(MOD_DIR) \
		-FindDir $(FIND_DIR) \
		-ClearDir $(CLEAR_DIR)

# DEV
aspire-run:
	$(Q)echo Starting NAUR stack with .NET Aspire...
	$(Q)dotnet run --project $(APPHOST_PRJ)

aspire-watch:
	$(Q)echo Starting NAUR stack with hot-reload...
	$(Q)dotnet watch --project $(APPHOST_PRJ) run

# MAINTENANCE & CLEANING
kill-dev:
	$(Q)$(PWSH) $(SCRIPTS_DIR)/kill-dev.ps1

clean:
	$(Q)$(PWSH) $(SCRIPTS_DIR)/clean.ps1

clean-all: clean
	$(Q)$(PWSH) $(SCRIPTS_DIR)/clean-aspire.ps1

# TROUBLESHOOTING
troubleshoot:
	$(Q)$(PWSH) $(SCRIPTS_DIR)/troubleshoot.ps1 \
		-AuthDir $(AUTH_DIR) \
		-NaurDir $(NAUR_DIR) \
		-ModDir $(MOD_DIR) \
		-FindDir $(FIND_DIR) \
		-ClearDir $(CLEAR_DIR) \
		-AppHostPrj $(APPHOST_PRJ)

# HELP
help:
	$(Q)$(PWSH) $(SCRIPTS_DIR)/help.ps1
