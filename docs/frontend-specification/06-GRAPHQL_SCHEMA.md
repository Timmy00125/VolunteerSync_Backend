# VolunteerSync Frontend Specification - GraphQL Schema Reference

## Document Information

- **Version**: 1.0
- **Date**: September 2, 2025
- **Status**: Draft
- **Previous Document**: [Testing Recommendations](./05-TESTING_RECOMMENDATIONS.md)

---

## 1. GraphQL Schema Overview

This document provides a comprehensive reference for the VolunteerSync GraphQL API schema. The schema defines the complete data model and operations available for the frontend application.

### 1.1 Schema Philosophy

- **Schema-First Design**: The API is designed with a schema-first approach using gqlgen
- **Type Safety**: Strong typing ensures data consistency across frontend and backend
- **Performance Optimized**: Designed to minimize N+1 queries and support efficient data fetching
- **Extensible**: Schema supports future enhancements while maintaining backward compatibility

### 1.2 Core Concepts

- **Pagination**: Relay-style cursor-based pagination for efficient data loading
- **Real-time Updates**: Subscription support for live updates (to be implemented)
- **Error Handling**: Structured error responses with detailed context
- **File Uploads**: Support for image and document uploads via Upload scalar

---

## 2. Scalar Types

### 2.1 Custom Scalars

```graphql
scalar Time # RFC3339 datetime format
scalar Upload # File upload type for multipart requests
scalar DateTime # Alternative datetime scalar for specific use cases
```

#### Usage Examples

```typescript
// Time scalar usage in TypeScript
interface Event {
  startTime: string; // "2024-03-15T09:00:00Z"
  endTime: string; // "2024-03-15T12:00:00Z"
}

// Upload scalar usage in Apollo
const UPLOAD_PROFILE_PICTURE = gql`
  mutation UploadProfilePicture($file: Upload!) {
    uploadProfilePicture(file: $file)
  }
`;
```

---

## 3. Core Types

### 3.1 User Type

**Purpose**: Represents authenticated users in the system

```graphql
type User {
  id: ID!
  email: String!
  name: String!
  emailVerified: Boolean!
  googleId: String
  lastLogin: Time
  createdAt: Time!
  updatedAt: Time!

  # Profile Information
  bio: String
  location: Location
  profilePicture: String
  interests: [Interest!]!
  skills: [Skill!]!
  roles: [String!]!
  isVerified: Boolean!
  joinedAt: Time!
  lastActiveAt: Time
  publicProfile: PublicProfile!
}
```

#### Frontend Implementation Notes

- **Caching**: User data should be cached globally for the session
- **Profile Updates**: Use optimistic updates for profile changes
- **Image Handling**: Profile pictures should support lazy loading and fallbacks

### 3.2 Event Type

**Purpose**: Core volunteer opportunity representation

```graphql
type Event {
  id: ID!
  title: String!
  description: String!
  shortDescription: String
  organizer: User!
  organizerId: ID!
  status: EventStatus!
  startTime: Time!
  endTime: Time!
  location: EventLocation!
  capacity: EventCapacity!
  requirements: EventRequirements!
  category: EventCategory!
  timeCommitment: TimeCommitmentType!
  tags: [String!]!
  slug: String
  shareURL: String
  recurrenceRule: RecurrenceRule
  registrationSettings: RegistrationSettings!
  images: [EventImage!]!
  announcements: [EventAnnouncement!]!
  createdAt: Time!
  updatedAt: Time!

  # Computed Fields
  currentRegistrations: Int!
  availableSpots: Int!
  isAtCapacity: Boolean!
  canRegister: Boolean!
}
```

#### Key Computed Fields

- **currentRegistrations**: Real-time count of confirmed registrations
- **availableSpots**: Remaining capacity (maximum - current)
- **isAtCapacity**: Boolean indicating if event is full
- **canRegister**: Whether current user can register (considers permissions, timing, capacity)

#### Frontend Implementation Notes

- **Data Loading**: Use DataLoader pattern to prevent N+1 queries for event lists
- **Real-time Updates**: Subscribe to capacity changes for live availability
- **SEO Optimization**: Use slug field for user-friendly URLs

### 3.3 Registration Type

**Purpose**: Manages volunteer registration for events

```graphql
type Registration {
  id: ID!
  user: User!
  event: Event!
  status: RegistrationStatus!
  personalMessage: String
  skills: [UserSkill!]!
  interests: [Interest!]!
  appliedAt: DateTime!
  confirmedAt: DateTime
  cancelledAt: DateTime
  checkedInAt: DateTime
  completedAt: DateTime
  waitlistPosition: Int
  approvalNotes: String
  cancellationReason: String
  attendanceStatus: AttendanceStatus!
  canCancel: Boolean!
  canCheckIn: Boolean!
  createdAt: DateTime!
  updatedAt: DateTime!
}
```

#### Status Flow

```
PENDING_APPROVAL → CONFIRMED → CHECKED_IN → COMPLETED
                → DECLINED
                → WAITLISTED → CONFIRMED
                → CANCELLED
```

---

## 4. Enums and Constants

### 4.1 Event Status

```graphql
enum EventStatus {
  DRAFT # Event is being created/edited
  PUBLISHED # Event is live and accepting registrations
  CANCELLED # Event has been cancelled
  COMPLETED # Event has concluded
  ARCHIVED # Event is archived for historical purposes
}
```

### 4.2 Event Categories

```graphql
enum EventCategory {
  COMMUNITY_SERVICE
  ENVIRONMENTAL
  EDUCATION
  HEALTH_WELLNESS
  DISASTER_RELIEF
  ANIMAL_WELFARE
  ARTS_CULTURE
  TECHNOLOGY
  SPORTS_RECREATION
  FOOD_HUNGER
  YOUTH_DEVELOPMENT
  SENIOR_CARE
  HOMELESS_SERVICES
  FUNDRAISING
  ADVOCACY
  OTHER
}
```

### 4.3 Registration Status

```graphql
enum RegistrationStatus {
  PENDING_APPROVAL # Awaiting organizer approval
  CONFIRMED # Registration confirmed
  WAITLISTED # On waitlist due to capacity
  CANCELLED # User cancelled registration
  DECLINED # Organizer declined registration
  NO_SHOW # User didn't attend
  COMPLETED # User completed the event
}
```

### 4.4 Time Commitment Types

```graphql
enum TimeCommitmentType {
  ONE_TIME # Single occurrence
  WEEKLY # Weekly recurring
  MONTHLY # Monthly recurring
  SEASONAL # Seasonal (quarterly/yearly)
  ONGOING # Long-term commitment
}
```

---

## 5. Query Operations

### 5.1 Authentication Queries

```graphql
type Query {
  # Get current authenticated user
  me: User

  # Get public profile by ID
  user(id: ID!): PublicProfile

  # Search users with filters
  searchUsers(
    filter: UserSearchFilter!
    limit: Int
    offset: Int
  ): [PublicProfile!]!
}
```

#### Frontend Usage Examples

```typescript
// Get current user
const GET_CURRENT_USER = gql`
  query GetCurrentUser {
    me {
      id
      name
      email
      profilePicture
      roles
      emailVerified
    }
  }
`;

// Search users
const SEARCH_USERS = gql`
  query SearchUsers($filter: UserSearchFilter!, $limit: Int) {
    searchUsers(filter: $filter, limit: $limit) {
      id
      name
      bio
      location {
        city
        state
      }
      skills {
        name
        proficiency
      }
    }
  }
`;
```

### 5.2 Event Queries

```graphql
type Query {
  # Get single event by ID
  event(id: ID!): Event

  # Get event by URL slug
  eventBySlug(slug: String!): Event

  # List events with pagination and filtering
  events(
    filter: EventSearchFilter
    sort: EventSortInput
    first: Int
    after: String
  ): EventConnection!

  # Search events by text query
  searchEvents(
    query: String!
    filter: EventSearchFilter
    sort: EventSortInput
    first: Int
    after: String
  ): EventConnection!

  # Get events organized by current user
  myEvents(status: [EventStatus!], first: Int, after: String): EventConnection!

  # Find nearby events by location
  nearbyEvents(
    coordinates: CoordinatesInput!
    radius: Float!
    filter: EventSearchFilter
    first: Int
    after: String
  ): EventConnection!
}
```

#### Event Filtering Examples

```typescript
// Filter events by category and date range
const GET_EVENTS = gql`
  query GetEvents($filter: EventSearchFilter, $first: Int, $after: String) {
    events(filter: $filter, first: $first, after: $after) {
      edges {
        node {
          id
          title
          shortDescription
          startTime
          endTime
          location {
            name
            city
            state
          }
          category
          currentRegistrations
          availableSpots
          canRegister
          images {
            url
            altText
            isPrimary
          }
        }
        cursor
      }
      pageInfo {
        hasNextPage
        endCursor
      }
      totalCount
    }
  }
`;

// Search events with text query
const SEARCH_EVENTS = gql`
  query SearchEvents($query: String!, $filter: EventSearchFilter) {
    searchEvents(query: $query, filter: $filter, first: 20) {
      edges {
        node {
          id
          title
          shortDescription
          category
          startTime
          location {
            name
            city
          }
        }
      }
    }
  }
`;
```

### 5.3 Registration Queries

```graphql
type Query {
  # Get user's registrations
  myRegistrations(filter: RegistrationFilterInput): [Registration!]!

  # Get specific registration
  registration(id: ID!): Registration

  # Get registrations for an event (organizer only)
  eventRegistrations(
    eventId: ID!
    filter: RegistrationFilterInput
  ): [Registration!]!

  # Get waitlist for an event
  waitlistEntries(eventId: ID!): [WaitlistEntry!]!

  # Check registration conflicts
  registrationConflicts(eventId: ID!): [RegistrationConflict!]!
}
```

---

## 6. Mutation Operations

### 6.1 Authentication Mutations

```graphql
type Mutation {
  # User registration
  register(input: RegisterInput!): AuthPayload!

  # User login
  login(input: LoginInput!): AuthPayload!

  # Refresh authentication token
  refreshToken(input: RefreshTokenInput!): AuthPayload!

  # User logout
  logout: Boolean!

  # Google OAuth flow
  googleAuthURL(redirectURL: String!): String!
  googleCallback(
    code: String!
    state: String!
    redirectURL: String!
  ): AuthPayload!
}
```

#### Authentication Flow Example

```typescript
// User registration
const REGISTER_USER = gql`
  mutation RegisterUser($input: RegisterInput!) {
    register(input: $input) {
      token
      refreshToken
      user {
        id
        name
        email
        emailVerified
      }
    }
  }
`;

// Google OAuth flow
const GET_GOOGLE_AUTH_URL = gql`
  mutation GetGoogleAuthURL($redirectURL: String!) {
    googleAuthURL(redirectURL: $redirectURL)
  }
`;

const GOOGLE_CALLBACK = gql`
  mutation GoogleCallback(
    $code: String!
    $state: String!
    $redirectURL: String!
  ) {
    googleCallback(code: $code, state: $state, redirectURL: $redirectURL) {
      token
      refreshToken
      user {
        id
        name
        email
        googleId
      }
    }
  }
`;
```

### 6.2 Profile Management Mutations

```graphql
type Mutation {
  # Update user profile
  updateProfile(input: UpdateProfileInput!): User!

  # Upload profile picture
  uploadProfilePicture(file: Upload!): String!

  # Manage interests and skills
  updateInterests(input: InterestInput!): User!
  addSkill(input: SkillInput!): User!
  removeSkill(skillId: ID!): User!

  # Privacy and preferences
  updatePrivacySettings(input: PrivacySettingsInput!): User!
  updateNotificationPreferences(input: NotificationPreferencesInput!): User!

  # Account management
  changePassword(currentPassword: String!, newPassword: String!): Boolean!
  deactivateAccount(confirmationCode: String!): Boolean!
  exportUserData: String!
}
```

### 6.3 Event Management Mutations

```graphql
type Mutation {
  # Event lifecycle
  createEvent(input: CreateEventInput!): Event!
  updateEvent(id: ID!, input: UpdateEventInput!): Event!
  publishEvent(id: ID!): Event!
  cancelEvent(id: ID!, reason: String): Event!
  deleteEvent(id: ID!): Boolean!

  # Event media management
  addEventImage(
    eventId: ID!
    file: Upload!
    altText: String
    isPrimary: Boolean
  ): EventImage!
  updateEventImage(
    id: ID!
    altText: String
    isPrimary: Boolean
    displayOrder: Int
  ): EventImage!
  deleteEventImage(id: ID!): Boolean!

  # Event announcements
  createEventAnnouncement(
    eventId: ID!
    title: String!
    content: String!
    isUrgent: Boolean
  ): EventAnnouncement!
}
```

### 6.4 Registration Mutations

```graphql
type Mutation {
  # Registration lifecycle
  registerForEvent(input: RegisterForEventInput!): Registration!
  cancelRegistration(registrationId: ID!, reason: String): Registration!
  updateRegistration(
    registrationId: ID!
    personalMessage: String
  ): Registration!

  # Bulk operations
  bulkRegister(input: BulkRegistrationInput!): [Registration!]!

  # Organizer operations
  approveRegistration(input: ApprovalDecisionInput!): Registration!
  promoteFromWaitlist(registrationId: ID!): Registration!

  # Attendance tracking
  checkInVolunteer(input: AttendanceInput!): AttendanceRecord!
  markAttendance(input: AttendanceInput!): AttendanceRecord!
}
```

#### Registration Flow Examples

```typescript
// Register for event
const REGISTER_FOR_EVENT = gql`
  mutation RegisterForEvent($input: RegisterForEventInput!) {
    registerForEvent(input: $input) {
      id
      status
      personalMessage
      appliedAt
      user {
        id
        name
      }
      event {
        id
        title
        startTime
      }
    }
  }
`;

// Cancel registration
const CANCEL_REGISTRATION = gql`
  mutation CancelRegistration($registrationId: ID!, $reason: String) {
    cancelRegistration(registrationId: $registrationId, reason: $reason) {
      id
      status
      cancelledAt
      cancellationReason
    }
  }
`;

// Bulk register for multiple events
const BULK_REGISTER = gql`
  mutation BulkRegister($input: BulkRegistrationInput!) {
    bulkRegister(input: $input) {
      id
      status
      event {
        id
        title
        startTime
      }
    }
  }
`;
```

---

## 7. Input Types

### 7.1 Event Creation Input

```graphql
input CreateEventInput {
  title: String!
  description: String!
  shortDescription: String
  startTime: Time!
  endTime: Time!
  location: EventLocationInput!
  capacity: EventCapacityInput!
  requirements: EventRequirementsInput
  tags: [String!]
  category: EventCategory!
  timeCommitment: TimeCommitmentType!
  recurrenceRule: RecurrenceRuleInput
  registrationSettings: RegistrationSettingsInput!
}

input EventLocationInput {
  name: String!
  address: String!
  city: String!
  state: String
  country: String!
  zipCode: String
  coordinates: CoordinatesInput
  instructions: String
  isRemote: Boolean!
}

input EventCapacityInput {
  minimum: Int!
  maximum: Int!
  waitlistEnabled: Boolean!
}
```

### 7.2 Search and Filter Inputs

```graphql
input EventSearchFilter {
  query: String
  status: [EventStatus!]
  category: [EventCategory!]
  timeCommitment: [TimeCommitmentType!]
  organizerId: ID
  tags: [String!]
  startDate: Time
  endDate: Time
  location: LocationSearchInput
  skills: [String!]
  interests: [String!]
  requiresBackgroundCheck: Boolean
  minimumAge: Int
}

input LocationSearchInput {
  city: String
  state: String
  country: String
  radius: Float
  coordinates: CoordinatesInput
}

input EventSortInput {
  field: EventSortField!
  direction: SortDirection!
}
```

### 7.3 Registration Inputs

```graphql
input RegisterForEventInput {
  eventId: ID!
  personalMessage: String
  emergencyContact: EmergencyContactInput
  dietaryRestrictions: String
  accessibilityNeeds: String
}

input EmergencyContactInput {
  name: String!
  phone: String!
}

input BulkRegistrationInput {
  eventIds: [ID!]!
  personalMessage: String
  skipConflicts: Boolean
}
```

---

## 8. Connection and Pagination Types

### 8.1 Relay-Style Pagination

```graphql
type EventConnection {
  edges: [EventEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type EventEdge {
  node: Event!
  cursor: String!
}

type PageInfo {
  hasNextPage: Boolean!
  hasPreviousPage: Boolean!
  startCursor: String
  endCursor: String
}
```

#### Frontend Pagination Implementation

```typescript
// Apollo Client pagination helper
const GET_EVENTS_WITH_PAGINATION = gql`
  query GetEventsWithPagination($first: Int!, $after: String) {
    events(first: $first, after: $after) {
      edges {
        node {
          id
          title
          startTime
          location {
            name
            city
          }
        }
        cursor
      }
      pageInfo {
        hasNextPage
        endCursor
      }
      totalCount
    }
  }
`;

// Apollo Client fetchMore implementation
const { data, loading, fetchMore } = useQuery(GET_EVENTS_WITH_PAGINATION, {
  variables: { first: 20 },
});

const loadMore = () => {
  if (data?.events.pageInfo.hasNextPage) {
    fetchMore({
      variables: {
        after: data.events.pageInfo.endCursor,
      },
      updateQuery: (prev, { fetchMoreResult }) => {
        if (!fetchMoreResult) return prev;

        return {
          events: {
            ...fetchMoreResult.events,
            edges: [...prev.events.edges, ...fetchMoreResult.events.edges],
          },
        };
      },
    });
  }
};
```

---

## 9. Error Handling

### 9.1 GraphQL Error Structure

```typescript
interface GraphQLError {
  message: string;
  locations?: Array<{
    line: number;
    column: number;
  }>;
  path?: Array<string | number>;
  extensions?: {
    code: string;
    timestamp: string;
    traceId?: string;
    details?: any;
  };
}
```

### 9.2 Common Error Codes

- **UNAUTHENTICATED**: User not logged in
- **UNAUTHORIZED**: User lacks permission
- **VALIDATION_ERROR**: Input validation failed
- **NOT_FOUND**: Resource not found
- **CONFLICT**: Resource conflict (e.g., event time overlap)
- **RATE_LIMITED**: Too many requests
- **INTERNAL_ERROR**: Server error

### 9.3 Frontend Error Handling

```typescript
// Error handling with Apollo Client
const [registerForEvent] = useMutation(REGISTER_FOR_EVENT, {
  onError: (error) => {
    if (error.graphQLErrors) {
      error.graphQLErrors.forEach(({ message, extensions }) => {
        switch (extensions?.code) {
          case "VALIDATION_ERROR":
            // Show field-specific validation errors
            showValidationErrors(extensions.details);
            break;
          case "CONFLICT":
            // Show conflict resolution options
            showConflictDialog(extensions.details);
            break;
          case "UNAUTHENTICATED":
            // Redirect to login
            router.navigate("/auth/login");
            break;
          default:
            // Show generic error message
            showErrorToast(message);
        }
      });
    }

    if (error.networkError) {
      // Handle network errors
      showNetworkErrorMessage();
    }
  },
});
```

---

## 10. Subscription Operations (Future)

### 10.1 Real-time Event Updates

```graphql
type Subscription {
  # Event capacity changes
  eventCapacityUpdated(eventId: ID!): Event!

  # Registration status updates
  registrationStatusChanged(userId: ID!): Registration!

  # New announcements
  eventAnnouncementAdded(eventId: ID!): EventAnnouncement!

  # Waitlist promotions
  waitlistPromoted(userId: ID!): Registration!
}
```

### 10.2 Frontend Subscription Implementation

```typescript
// Subscribe to event capacity changes
const EVENT_CAPACITY_SUBSCRIPTION = gql`
  subscription EventCapacityUpdated($eventId: ID!) {
    eventCapacityUpdated(eventId: $eventId) {
      id
      currentRegistrations
      availableSpots
      isAtCapacity
      canRegister
    }
  }
`;

// React hook for subscription
const useEventCapacitySubscription = (eventId: string) => {
  const { data, loading } = useSubscription(EVENT_CAPACITY_SUBSCRIPTION, {
    variables: { eventId },
  });

  return {
    event: data?.eventCapacityUpdated,
    loading,
  };
};
```

---

## 11. Performance Considerations

### 11.1 Query Optimization

- **Field Selection**: Only request needed fields to minimize payload
- **Fragment Usage**: Use GraphQL fragments for reusable field selections
- **Query Batching**: Batch multiple queries in a single request
- **DataLoader**: Backend uses DataLoader to prevent N+1 queries

### 11.2 Caching Strategy

```typescript
// Apollo Client cache configuration
const cache = new InMemoryCache({
  typePolicies: {
    Event: {
      fields: {
        registrations: {
          merge(existing = [], incoming) {
            return [...existing, ...incoming];
          },
        },
      },
    },
    Query: {
      fields: {
        events: relayStylePagination(),
        myRegistrations: {
          merge(existing = [], incoming) {
            return incoming;
          },
        },
      },
    },
  },
});
```

### 11.3 Bundle Size Optimization

- **Fragments**: Use fragments to share field selections
- **Code Splitting**: Split GraphQL operations by feature
- **Query Complexity**: Monitor and limit query complexity

---

## 12. Development Tools

### 12.1 GraphQL Code Generation

```typescript
// graphql.config.js
module.exports = {
  schema: "./schema.graphql",
  documents: "./src/**/*.graphql",
  generates: {
    "./src/generated/graphql.tsx": {
      plugins: [
        "typescript",
        "typescript-operations",
        "typescript-react-apollo",
      ],
      config: {
        withHooks: true,
        withComponent: false,
        withHOC: false,
      },
    },
  },
};
```

### 12.2 Schema Validation

```typescript
// Schema validation in tests
import { buildSchema, validate } from "graphql";
import { queries } from "./queries";

describe("GraphQL Schema Validation", () => {
  const schema = buildSchema(schemaString);

  it("should validate all queries", () => {
    queries.forEach((query) => {
      const errors = validate(schema, query);
      expect(errors).toHaveLength(0);
    });
  });
});
```

---

## Implementation Priority

### Phase 1: Core Functionality (Weeks 1-4)

- [ ] Authentication queries and mutations
- [ ] Basic event queries with pagination
- [ ] User profile management
- [ ] Event registration flow

### Phase 2: Advanced Features (Weeks 5-8)

- [ ] Advanced event filtering and search
- [ ] Bulk operations and conflict resolution
- [ ] Event management for organizers
- [ ] Waitlist and attendance tracking

### Phase 3: Optimization (Weeks 9-12)

- [ ] Real-time subscriptions
- [ ] Advanced caching strategies
- [ ] Performance monitoring
- [ ] Error tracking and analytics

---

**Schema Documentation**: This reference should be updated whenever the GraphQL schema changes. Frontend developers should regenerate TypeScript types after schema updates to maintain type safety.
