# CLAUDE.md

This file provides guidance to Claude Code when working with this repository.

## Project Overview

Open Workflow is an open source no-code workflow engine written in Go.

## Commands

- `make test` - Run all tests
- `make build` - Build the binary to `bin/api`
- `make run` - Run the application (serves on port 3000)
- `make fmt` - Format all Go files
- `make clean` - Remove build artifacts

## Architecture

This project follows hexagonal (ports and adapters) architecture with DDD patterns:

- `cmd/api/` - Application entry point
- `api/` - HTTP handlers and routing (Fiber framework)
- `di/` - Dependency injection container
- `internal/`
  - `domain/` - Domain aggregates, events, and business logic
  - `port/` - Port interfaces (inbound services, outbound repositories)
  - `adapter/` - Adapter implementations
- `pkg/` - Shared packages (domain base types, ID generation)

## Code Conventions

- Use Go standard formatting (`go fmt`)
- Follow hexagonal architecture: domain logic in `internal/domain/`, ports in `internal/port/`, adapters in `internal/adapter/`
- Aggregates contain domain events and business rules
- DTOs for inbound ports, models for outbound ports
- Mappers handle conversion between layers
