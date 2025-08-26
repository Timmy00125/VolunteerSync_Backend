# GitHub Copilot Instructions - VolunteerSync Backend

## Project Overview

**VolunteerSync** backend: Modern volunteer management platform using Go, Gin, and GraphQL. Must be fast, scalable, and modular for thousands of concurrent users.

## Core Architecture

### Tech Stack

- **Go 1.21+** with Gin framework
- **GraphQL** via gqlgen (schema-first)
- **PostgreSQL** with Redis caching
- **JWT authentication** with refresh tokens

### Key Principles

- Performance-first design
- Modular, loosely-coupled components
- Horizontal scalability
- Security by design

## Project Structure

The project follows a layered architecture with clear separation of concerns:

- **cmd/api/main.go** - Application entry point and server initialization
- **internal/graph/** - GraphQL layer with thin resolvers that delegate to services
- **internal/core/** - Business logic services organized by domain (auth, user, event)
- **internal/store/** - Data access layer with interfaces for testability
- **internal/middleware/** - HTTP middleware for authentication, logging, CORS
- **database/migrations/** - Version-controlled database schema changes
- **tests/** - Integration tests that verify end-to-end functionality

## Performance Standards

### Database Optimization

- Always use proper database indexing for frequently queried columns
- Implement connection pooling to handle concurrent requests efficiently
- Use batch operations with ANY() clauses instead of loops to reduce query count
- Prefer single queries over multiple round trips when possible

### GraphQL Optimization

- Use DataLoader pattern to prevent N+1 query problems in nested resolvers
- Implement field-level authorization to avoid unnecessary data fetching
- Cache resolver results when data doesn't change frequently

### Caching Strategy

- Cache frequently accessed data in Redis with database fallback
- Use cache invalidation patterns to maintain data consistency
- Implement cache warming for predictable access patterns

## Code Patterns

### Service Layer Design

- Each service should have a clear interface defining its contract
- Use constructor functions for dependency injection and initialization
- Services should only depend on interfaces, not concrete implementations
- Keep business logic separate from GraphQL resolvers

### Error Handling

- Use custom error types for different categories of errors
- Wrap errors with additional context using fmt.Errorf with %w verb
- Return structured errors that can be properly handled by GraphQL layer
- Log errors with sufficient context for debugging

### State Management

- Make all services stateless to support horizontal scaling
- Store session data in Redis, not in-memory structures
- Use database transactions for operations that must be atomic
- Implement proper cleanup for background processes

## Critical Anti-patterns to Avoid

### N+1 Query Problems

The most common GraphQL performance issue is executing one database query per item in a list. Instead of querying registrations individually for each event, use DataLoaders to batch these queries into a single database call.

### Blocking HTTP Requests

Never perform slow operations like sending emails or file processing directly in HTTP request handlers. These operations should be queued as background jobs to maintain response time targets.

### Memory Leaks in Long-Running Services

Avoid accumulating state in service structs or global variables. Use proper context cancellation and cleanup goroutines that could run indefinitely.

### Insecure Data Access

Always validate user permissions before returning data. Don't rely solely on GraphQL field-level permissions - implement authorization checks in the service layer.

## Performance Targets

- API response time under 200ms for 95th percentile
- Database query execution under 50ms for 95th percentile
- Support for 1000+ concurrent users without degradation
- Test coverage above 90% for business logic

## Development Workflow

### Code Generation

- Regenerate GraphQL resolvers and types after schema changes
- Run go generate to update all auto-generated code
- Verify generated code compiles before committing

### Database Management

- Use migrations for all database schema changes
- Test migrations both up and down directions
- Never modify existing migrations - create new ones for changes

### Code Quality

- Format code with gofmt and run golangci-lint before commits
- Write unit tests for all business logic functions
- Create integration tests for GraphQL mutations and queries

**Focus**: Every implementation should prioritize performance, maintain modularity, and scale horizontally. Write fast, secure, and maintainable code that follows established patterns.
