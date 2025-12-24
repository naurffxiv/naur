# Authingway
[![Build](https://github.com/naurffxiv/authingway/actions/workflows/build-code.yml/badge.svg?branch=main&event=push)](https://github.com/naurffxiv/authingway/actions/workflows/build-code.yml)
[![Lint](https://github.com/naurffxiv/authingway/actions/workflows/lint-code.yml/badge.svg?branch=main&event=push)](https://github.com/naurffxiv/authingway/actions/workflows/lint-code.yml)
[![Test](https://github.com/naurffxiv/authingway/actions/workflows/test-code.yml/badge.svg?branch=main&event=push)](https://github.com/naurffxiv/authingway/actions/workflows/test-code.yml)
[![Scan](https://github.com/naurffxiv/authingway/actions/workflows/scan-code.yml/badge.svg?branch=main&event=push)](https://github.com/naurffxiv/authingway/actions/workflows/scan-code.yml)

## What is Authingway?
Authingway is the auth service for the NAUR Discord server.

The goal is to provide a centralized auth service for NAUR developers to work with for user management as well as login for non-dioscord services.

## Getting started
To start developing Authingway you need just a few things:

- [.NET 10 SDK](https://dotnet.microsoft.com/en-us/download/dotnet/10.0)
- [Aspire CLI](https://aspire.dev/get-started/install-cli/)
- [Docker Desktop](https://www.docker.com/products/docker-desktop/), [Podman Desktop](https://podman-desktop.io/), or [Rancher Desktop](https://rancherdesktop.io/)

## Run Authingway
Once you have all the dependencies listed above simply execute `aspire run` to start Authingway. You can access the following services:

- Aspire Dashboard: https://aspire.dev.localhost:8440
- Authingway: https://authingway.dev.localhost:8443

The Aspire dashboard may prompt for additional values in order to start Authingway, filling these out will start the service.
