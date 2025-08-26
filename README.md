# VolunteerSync - A Modern Volunteer Management Platform

## 🎯 Project Overview

**VolunteerSync** is a comprehensive volunteer management platform designed to connect passionate volunteers with meaningful opportunities. Built with a modern, modular backend using the **Gin Web Framework** and **GraphQL** for Go, and a cross-platform Flutter application, it provides a seamless experience for both volunteers and event organizers.

### Core Mission

To connect dedicated volunteers with impactful opportunities while providing organizers with powerful, intuitive tools to manage events, track engagement, and build thriving communities.

---

## 🏗️ System Architecture

### High-Level Architecture

The application is built as a modular monolith using the Gin framework, exposing a flexible GraphQL API.

```
┌─────────────────┐      ┌──────────────────────────────┐
│   Flutter App   │◄────►│      Go Gin Backend          │
│  (Mobile/Web)   │      │   (GraphQL API on Port 8080) │
└─────────────────┘      └──────────────────────────────┘
                                     │
                                     │
                         ┌───────────▼────────────┐
                         │   PostgreSQL Database  │
                         │ (Single Source of Truth)│
                         └────────────────────────┘
```

### Component Responsibilities

#### 🎨 **Flutter Client Application**

- **Purpose**: A single, cross-platform mobile and web client for a consistent user experience.
- **Tech Stack**: Flutter/Dart.
- **Key Features**: User authentication, event discovery, registration management, real-time notifications, and personalized dashboards.
- **API Communication**: Interacts directly with the Go/Gin backend via a single GraphQL endpoint.

#### 🔧 **Go (Gin) Backend**

- **Purpose**: A single, modular backend service that handles all business logic, data processing, and the GraphQL API.
- **Tech Stack**: Go, Gin Web Framework, gqlgen (for GraphQL).
- **Key Responsibilities**:
  - Provides a secure and strongly-typed GraphQL API for the client application.
  - Manages user authentication and authorization using JWTs.
  - Handles all business logic for users, events, registrations, and more.
  - Interfaces directly with the PostgreSQL database.
- **Modular Structure**: The backend is organized into logical packages (e.g., `auth`, `graph`, `core`) to maintain separation of concerns and high cohesion.

#### 🗄️ **PostgreSQL Database**

- **Purpose**: The primary and single data store for the entire application.
- **Design**: A well-structured relational schema with clear table definitions, relationships, and performance-optimized indexes.
- **Key Features**: ACID compliance, UUID primary keys for all entities, and `TIMESTAMPTZ` for all date/time fields.

---

## 🎯 Core Features & User Stories

### 👤 **User Management**

- **Capabilities**: Multi-provider authentication (email/password, Google), profile management for volunteers and organizers, role-based access control.

### 🙋‍♀️ **Volunteer Experience**

- **Key Features**: Smart event discovery and filtering, one-click registration, personalized dashboards, real-time notifications, and achievement tracking.

### 🏢 **Organizer Tools**

- **Key Features**: End-to-end event management, volunteer roster management, attendance tracking, a central communication hub, and an analytics dashboard.

_(User stories remain the same as the original document)_

---

## 🔄 Key User Flows

### 🚀 **Volunteer Registration & Event Sign-up Flow**

1.  **Onboarding**: A new user opens the Flutter app and selects "Sign Up with Google."
2.  **Authentication**: The app communicates with the Gin backend's `/auth/google/callback` endpoint. The backend handles the OAuth2.0 flow, creates a new user record, and returns a JWT to the client.
3.  **Discovery**: The authenticated user browses events by sending a GraphQL `query` to the `/graphql` endpoint.
    ```graphql
    query GetEvents {
      events(filter: { published: true }) {
        id
        title
        startTime
        location
      }
    }
    ```
4.  **Registration**: The user clicks "Register" for an event, triggering a GraphQL `mutation`.
    ```graphql
    mutation RegisterForEvent($eventId: ID!) {
      registerForEvent(eventId: $eventId) {
        success
        registration {
          id
          status
        }
      }
    }
    ```

### 🏗️ **Organizer Event Creation Flow**

1.  **Access**: An organizer logs in and navigates to the "Create Event" screen in the app.
2.  **Creation**: The organizer fills out the event form. Submitting the form sends a `mutation` with the event data to the `/graphql` endpoint.
    ```graphql
    mutation CreateEvent($input: NewEvent!) {
      createEvent(input: $input) {
        id
        title
        description
      }
    }
    ```
3.  **Processing**: The Gin backend's GraphQL resolver validates the incoming data, creates a new event record in the database, and returns the newly created event object.
4.  **Management**: The organizer can then view registered volunteers by sending a `query`.
    ```graphql
    query GetEventRegistrations($eventId: ID!) {
      event(id: $eventId) {
        registrations {
          id
          user {
            id
            name
            email
          }
        }
      }
    }
    ```

---

## 📋 System Requirements

_(Functional and Non-Functional Requirements remain the same)_

---

## 🗄️ Database Schema Design

The PostgreSQL database schema remains the same, providing a solid foundation for the application.

_(The SQL schema definition from the original document remains unchanged here)_

---

## ⚙️ Backend Architecture (Gin + GraphQL)

The backend is a modular monolithic application built with Gin. It exposes a single `/graphql` endpoint, with `gqlgen` used to manage schema and resolvers.

### **Project Structure (gqlgen)**

```
volunteer-sync-backend/
├── cmd/
│   └── api/
│       └── main.go            # Application entry point
├── internal/
│   ├── graph/                 # GraphQL specific code
│   │   ├── generated/
│   │   │   └── generated.go   # Auto-generated by gqlgen
│   │   ├── model/
│   │   │   └── models_gen.go  # Auto-generated models from schema
│   │   ├── resolver/
│   │   │   ├── resolver.go    # Root resolver implementation
│   │   │   ├── schema.resolvers.go # Generated resolver stubs
│   │   │   └── user.resolvers.go # Example resolver for User type
│   │   └── schema.graphqls      # GraphQL schema definition
│   ├── core/                  # Business logic services
│   │   ├── auth_service.go
│   │   ├── user_service.go
│   │   └── event_service.go
│   ├── store/                 # Database interaction (repositories)
│   │   ├── postgres/
│   │   │   ├── user_store.go
│   │   │   └── event_store.go
│   │   └── models.go          # GORM or sqlx models
│   ├── config/                # Configuration management
│   └── pkg/                   # Shared packages (e.g., jwt, validation)
├── database/
│   └── migrations/            # Database migration files
├── go.mod
├── go.sum
└── Dockerfile
```

### **GraphQL Schema Highlights**

This section defines the primary GraphQL types, queries, and mutations.

#### **Core Types**

```graphql
type User {
  id: ID!
  name: String!
  email: String!
  roles: [String!]!
  interests: [String!]
}

type Event {
  id: ID!
  title: String!
  description: String!
  startTime: Time!
  endTime: Time!
  location: String!
  organizer: User!
  registrations: [Registration!]
}

type AuthPayload {
  accessToken: String!
  refreshToken: String!
  user: User!
}
```

#### **Queries**

```graphql
type Query {
  # Get a list of all published events with filtering and pagination.
  events(filter: EventFilter, limit: Int, offset: Int): [Event!]!

  # Get the details of a specific event.
  event(id: ID!): Event

  # Get the profile of the currently authenticated user.
  me: User

  # Get the public profile of a specific user.
  user(id: ID!): User
}
```

#### **Mutations**

```graphql
type Mutation {
  # Register a new user with email and password.
  register(input: RegisterInput!): AuthPayload!

  # Log in a user and return a JWT.
  login(input: LoginInput!): AuthPayload!

  # Refresh an expired access token using a refresh token.
  refreshToken(token: String!): AuthPayload!

  # (Organizer only) Create a new event.
  createEvent(input: NewEvent!): Event!

  # (Organizer only) Update an event.
  updateEvent(id: ID!, input: UpdateEvent!): Event!

  # Register the current user for an event.
  registerForEvent(eventId: ID!): Registration!

  # Cancel the current user's registration for an event.
  cancelRegistration(eventId: ID!): Boolean!
}
```

### **Resolver and Service Signatures**

**GraphQL Resolver Example (`internal/graph/resolver/schema.resolvers.go`)**

```go
// Register handles the 'register' mutation.
func (r *mutationResolver) Register(ctx context.Context, input model.RegisterInput) (*model.AuthPayload, error) {
    // The resolver's job is to map the GraphQL request to the correct service call.
    // It should contain minimal business logic.

    // Call the business logic service
    authResponse, err := r.AuthService.Register(ctx, &input)
    if err != nil {
        // Handle specific errors (e.g., user already exists)
        return nil, err // gqlgen handles error formatting
    }

    return authResponse, nil
}
```

**Service Logic Example (`internal/core/auth_service.go`)**

```go
// Register creates a new user and returns an authentication response.
func (s *AuthService) Register(ctx context.Context, req *model.RegisterInput) (*model.AuthPayload, error) {
    // 1. Validate input
    // 2. Check if user already exists in the database
    // 3. Hash the password
    // 4. Create the user record
    // 5. Generate JWT access and refresh tokens
    // 6. Return the AuthPayload
    return &model.AuthPayload{...}, nil
}
```

---

## 🔐 Authentication & Security Strategy

- **JWT-based Authentication**: The Gin backend uses a middleware to protect the `/graphql` endpoint. The middleware inspects the `Authorization: Bearer <token>` header, validates the JWT, and extracts user claims. The user information is then added to the `context.Context` for use in GraphQL resolvers.
- **Secure Token Handling**: The client is responsible for securely storing the JWT (e.g., in secure storage) and sending it with every request to the GraphQL API.
- **Password Security**: Passwords are hashed using `bcrypt` with a cost factor of at least 12.
- **CORS**: The Gin backend is configured with a strict Cross-Origin Resource Sharing (CORS) policy to only allow requests from the Flutter web app's domain in production.

---

## 🛠️ Development Guidelines & Best Practices

### **Go (Gin + gqlgen) Development Standards**

- **Schema-First Development**: Define your API in `schema.graphqls` first. Use `go run github.com/99designs/gqlgen generate` to create models and resolver stubs.
- **Project Structure**: Follow the modular structure outlined above, keeping resolvers thin and business logic in the `core` services.
- **Error Handling**: Return errors directly from resolvers. `gqlgen` will format them into a standard GraphQL error response. Use custom error types for specific scenarios (e.g., `ErrNotFound`, `ErrPermissionDenied`).
- **Context Propagation**: Pass the `context.Context` from resolvers down to the service and repository layers for request-scoped values, cancellation, and timeouts.
- **Testing**: Write unit tests for services and repository methods. Write integration tests for the GraphQL API by constructing and sending queries to a test server.
- **Configuration**: Manage configuration using a library like Viper, loading values from environment variables to adhere to 12-factor app principles.

_(Flutter and DevOps guidelines remain the same)_

---

## 📝 Development Priorities & Roadmap

_(Roadmap remains the same)_

---

## 🎯 Success Metrics & KPIs

_(KPIs remain the same)_

---

_This document serves as the comprehensive development guide for the VolunteerSync platform. It should be referenced regularly to ensure consistency and quality._
