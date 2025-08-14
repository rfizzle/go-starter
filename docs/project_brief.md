# Project Brief

---

## 1. Project Overview

### 1.1. Project Purpose

This Go Starter project provides a production-ready, enterprise-grade foundation for building scalable REST API services
in Go. It eliminates the common overhead of setting up basic infrastructure, authentication, database integration, and
observability features that every modern web service requires. The starter accelerates development by providing a
well-architected, tested foundation that follows Go best practices and industry standards.

### 1.2. Goals

*   **Goal 1:** Provide a complete, production-ready Go web service foundation with minimal configuration required
*   **Goal 2:** Implement OpenAPI-first development workflow with automatic code generation for consistent API development
*   **Goal 3:** Deliver enterprise-grade features including authentication, rate limiting, observability, and database abstraction out of the box

---

## 2. Business Requirements

### 2.1. Functional Requirements

*   The system must provide health check endpoints (`/healthz/liveness`, `/healthz/readiness`) for Kubernetes deployment compatibility
*   The system must implement JWT-based authentication with RBAC (Role-Based Access Control) capabilities
*   The system must provide a complete PostgreSQL integration layer with migration support and transaction management
*   The system must expose Prometheus metrics and OpenTelemetry tracing for observability
*   The system must implement configurable rate limiting with multiple strategies
*   The system must support OpenAPI specification-driven API development with automatic code generation

### 2.2. Non-Functional Requirements

*   **Performance:** HTTP middleware stack optimized for production with configurable timeouts, request ID tracking, and recovery mechanisms
*   **Scalability:** Stateless design with database connection pooling and horizontal scaling capabilities
*   **Security:** Built-in authentication middleware, audit logging, CORS configuration, and security best practices implementation
*   **Reliability:** Comprehensive error handling, structured logging with Zap, and production-ready middleware chain
*   **Maintainability:** Clear separation of concerns with layered architecture, comprehensive test coverage, and automated code generation

---

## 3. Success Metrics

*   **Metric 1:** Development velocity—Reduce time to API implementation by 80% compared to starting from scratch
*   **Metric 2:** Code quality—Maintain >90% test coverage across all non-generated code
*   **Metric 3:** Production readiness—Zero additional configuration required for basic deployment to Kubernetes environments

---

## 4. Key Features (Implemented Foundation)

*   **OpenAPI-First Development:** Complete ogen-based code generation from OpenAPI specifications with automatic interface generation
*   **Authentication & Authorization:** JWT-based auth system with middleware, RBAC, audit logging, and context management
*   **Database Integration:** PostgreSQL abstraction layer with pgx/v5 driver, migration system, transaction management, and testing utilities  
*   **Web Server Framework:** Chi-based HTTP server with production middleware stack (timeouts, logging, CORS, static files, recovery)
*   **Observability Stack:** Prometheus metrics, OpenTelemetry tracing, pprof debugging endpoints, and structured logging
*   **Rate Limiting:** Configurable rate limiting middleware with multiple strategies and customization options
*   **CLI Interface:** Cobra-based command structure with Viper configuration management supporting flags and environment variables
*   **Development Tooling:** Comprehensive Makefile with build, test, lint, format, and generation commands for streamlined development workflow

---

## 5. Out of Scope

*   Frontend user interfaces or web applications—this is purely a backend API foundation
*   Specific business logic implementation—users must implement their own domain-specific controllers and models
*   Database vendor lock-in beyond PostgreSQL—other databases would require additional driver implementations
*   Real-time features like WebSockets or Server-Sent Events
*   Message queuing or event streaming capabilities
*   Third-party service integrations (payment processors, external APIs, etc.)
*   Container orchestration beyond basic Kubernetes health checks
*   Advanced deployment strategies (blue-green, canary deployments)