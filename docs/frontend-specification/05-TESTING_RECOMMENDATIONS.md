# VolunteerSync Frontend Specification - Testing Recommendations

## Document Information

- **Version**: 1.0
- **Date**: September 2, 2025
- **Status**: Draft
- **Previous Document**: [User Stories](./04-USER_STORIES.md)
- **Next Document**: [GraphQL Schema Reference](./06-GRAPHQL_SCHEMA.md)

---

## 1. Testing Strategy Overview

### 1.1 Testing Philosophy

The VolunteerSync frontend testing strategy follows the **Testing Pyramid** principle, emphasizing:

- **Unit Tests (70%)**: Fast, isolated tests for components and services
- **Integration Tests (20%)**: API integration and component interaction tests
- **End-to-End Tests (10%)**: Critical user journey validation

### 1.2 Quality Gates

- **Unit Test Coverage**: Minimum 90% for business logic
- **Integration Test Coverage**: Minimum 80% for API interactions
- **E2E Test Coverage**: 100% of critical user paths
- **Performance Tests**: Core Web Vitals compliance
- **Accessibility Tests**: WCAG 2.1 AA compliance

### 1.3 Testing Tools Stack

- **Unit Testing**: Jest + Angular Testing Utilities
- **Component Testing**: Angular TestBed + Testing Library
- **API Testing**: Apollo Testing utilities + MSW (Mock Service Worker)
- **E2E Testing**: Cypress
- **Visual Regression**: Percy or Chromatic
- **Performance Testing**: Lighthouse CI
- **Accessibility Testing**: axe-core + Pa11y

---

## 2. Unit Testing

### 2.1 Component Testing

**Scope**: Individual Angular components  
**Coverage Target**: 95%

#### Test Categories

```typescript
// Example: Event Card Component Test
describe("EventCardComponent", () => {
  let component: EventCardComponent;
  let fixture: ComponentFixture<EventCardComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [EventCardComponent],
      imports: [MaterialModule, RouterTestingModule],
      providers: [MockEventService],
    });
    fixture = TestBed.createComponent(EventCardComponent);
    component = fixture.componentInstance;
  });

  describe("Component Rendering", () => {
    it("should display event title and description", () => {
      // Given
      component.event = mockEventData;

      // When
      fixture.detectChanges();

      // Then
      expect(screen.getByText(mockEventData.title)).toBeInTheDocument();
      expect(screen.getByText(mockEventData.description)).toBeInTheDocument();
    });

    it("should show registration button for available events", () => {
      // Given
      component.event = {
        ...mockEventData,
        status: "PUBLISHED",
        canRegister: true,
      };

      // When
      fixture.detectChanges();

      // Then
      expect(
        screen.getByRole("button", { name: /register/i })
      ).toBeInTheDocument();
    });

    it("should show waitlist button when event is full", () => {
      // Given
      component.event = { ...mockEventData, isAtCapacity: true };

      // When
      fixture.detectChanges();

      // Then
      expect(
        screen.getByRole("button", { name: /join waitlist/i })
      ).toBeInTheDocument();
    });
  });

  describe("User Interactions", () => {
    it("should emit register event when register button clicked", () => {
      // Given
      spyOn(component.registerEvent, "emit");
      component.event = mockEventData;
      fixture.detectChanges();

      // When
      const registerButton = screen.getByRole("button", { name: /register/i });
      fireEvent.click(registerButton);

      // Then
      expect(component.registerEvent.emit).toHaveBeenCalledWith(mockEventData);
    });
  });

  describe("Edge Cases", () => {
    it("should handle missing event data gracefully", () => {
      // Given
      component.event = null;

      // When/Then
      expect(() => fixture.detectChanges()).not.toThrow();
    });

    it("should display placeholder when image fails to load", () => {
      // Test image fallback behavior
    });
  });
});
```

#### Component Testing Checklist

- [ ] **Rendering Tests**: Component displays correctly with various props
- [ ] **Interaction Tests**: User interactions trigger expected behavior
- [ ] **State Management**: Component state changes work correctly
- [ ] **Event Emission**: Output events are emitted with correct data
- [ ] **Error Handling**: Component handles invalid/missing data gracefully
- [ ] **Accessibility**: ARIA attributes and keyboard navigation work

### 2.2 Service Testing

**Scope**: Angular services and business logic  
**Coverage Target**: 95%

#### Service Test Example

```typescript
// Example: Event Service Test
describe("EventService", () => {
  let service: EventService;
  let apolloSpy: jasmine.SpyObj<Apollo>;

  beforeEach(() => {
    const spy = jasmine.createSpyObj("Apollo", ["watchQuery", "mutate"]);
    TestBed.configureTestingModule({
      providers: [EventService, { provide: Apollo, useValue: spy }],
    });
    service = TestBed.inject(EventService);
    apolloSpy = TestBed.inject(Apollo) as jasmine.SpyObj<Apollo>;
  });

  describe("getEvents", () => {
    it("should return events observable", () => {
      // Given
      const mockEvents = [mockEventData];
      apolloSpy.watchQuery.and.returnValue({
        valueChanges: of({ data: { events: mockEvents }, loading: false }),
      });

      // When
      const result$ = service.getEvents();

      // Then
      result$.subscribe((events) => {
        expect(events).toEqual(mockEvents);
      });
    });

    it("should handle GraphQL errors", () => {
      // Test error handling
    });
  });

  describe("registerForEvent", () => {
    it("should call registration mutation with correct variables", () => {
      // Test mutation calls
    });

    it("should update cache after successful registration", () => {
      // Test cache updates
    });
  });
});
```

#### Service Testing Checklist

- [ ] **Method Functionality**: All public methods work as expected
- [ ] **Error Handling**: Service handles API errors gracefully
- [ ] **Data Transformation**: Data is correctly transformed/mapped
- [ ] **Cache Management**: Apollo cache is updated correctly
- [ ] **Side Effects**: External calls and state changes are tested

### 2.3 Pipe and Directive Testing

**Scope**: Custom pipes and directives  
**Coverage Target**: 100%

#### Pipe Test Example

```typescript
describe("TimeAgoPipe", () => {
  let pipe: TimeAgoPipe;

  beforeEach(() => {
    pipe = new TimeAgoPipe();
  });

  it('should format recent dates as "just now"', () => {
    const now = new Date();
    expect(pipe.transform(now)).toBe("just now");
  });

  it("should format dates from yesterday", () => {
    const yesterday = new Date();
    yesterday.setDate(yesterday.getDate() - 1);
    expect(pipe.transform(yesterday)).toBe("1 day ago");
  });

  it("should handle null input gracefully", () => {
    expect(pipe.transform(null)).toBe("");
  });
});
```

---

## 3. Integration Testing

### 3.1 GraphQL Integration Testing

**Scope**: Apollo client integration and GraphQL operations  
**Coverage Target**: 90%

#### GraphQL Test Setup

```typescript
// GraphQL Integration Test Setup
describe("EventService GraphQL Integration", () => {
  let service: EventService;
  let controller: ApolloTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [ApolloTestingModule],
      providers: [EventService],
    });
    service = TestBed.inject(EventService);
    controller = TestBed.inject(ApolloTestingController);
  });

  afterEach(() => {
    controller.verify();
  });

  it("should fetch events with correct GraphQL query", () => {
    // Given
    const mockEventsResponse = {
      data: {
        events: {
          edges: [{ node: mockEventData }],
          pageInfo: { hasNextPage: false },
        },
      },
    };

    // When
    service.getEvents().subscribe((events) => {
      expect(events).toEqual([mockEventData]);
    });

    // Then
    const op = controller.expectOne(GET_EVENTS_QUERY);
    expect(op.operation.variables).toEqual({});
    op.flush(mockEventsResponse);
  });

  it("should handle registration mutation", () => {
    // Test mutation operations
  });

  it("should handle subscription for real-time updates", () => {
    // Test subscriptions
  });
});
```

#### GraphQL Testing Checklist

- [ ] **Query Operations**: All queries execute with correct variables
- [ ] **Mutation Operations**: Mutations update data and cache correctly
- [ ] **Subscription Operations**: Real-time updates work properly
- [ ] **Error Handling**: GraphQL errors are handled gracefully
- [ ] **Cache Updates**: Apollo cache is updated after mutations
- [ ] **Optimistic Updates**: UI updates optimistically for better UX

### 3.2 Route and Guard Testing

**Scope**: Angular routing and route guards  
**Coverage Target**: 100%

#### Route Guard Test Example

```typescript
describe("AuthGuard", () => {
  let guard: AuthGuard;
  let authService: jasmine.SpyObj<AuthService>;
  let router: jasmine.SpyObj<Router>;

  beforeEach(() => {
    const authSpy = jasmine.createSpyObj("AuthService", ["isAuthenticated"]);
    const routerSpy = jasmine.createSpyObj("Router", ["navigate"]);

    TestBed.configureTestingModule({
      providers: [
        AuthGuard,
        { provide: AuthService, useValue: authSpy },
        { provide: Router, useValue: routerSpy },
      ],
    });

    guard = TestBed.inject(AuthGuard);
    authService = TestBed.inject(AuthService) as jasmine.SpyObj<AuthService>;
    router = TestBed.inject(Router) as jasmine.SpyObj<Router>;
  });

  it("should allow access for authenticated users", () => {
    // Given
    authService.isAuthenticated.and.returnValue(true);

    // When
    const result = guard.canActivate();

    // Then
    expect(result).toBe(true);
  });

  it("should redirect to login for unauthenticated users", () => {
    // Given
    authService.isAuthenticated.and.returnValue(false);

    // When
    const result = guard.canActivate();

    // Then
    expect(result).toBe(false);
    expect(router.navigate).toHaveBeenCalledWith(["/auth/login"]);
  });
});
```

### 3.3 Form Integration Testing

**Scope**: Complex forms with validation  
**Coverage Target**: 95%

#### Form Test Example

```typescript
describe("EventRegistrationForm Integration", () => {
  let component: EventRegistrationFormComponent;
  let fixture: ComponentFixture<EventRegistrationFormComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [EventRegistrationFormComponent],
      imports: [ReactiveFormsModule, MaterialModule],
      providers: [MockRegistrationService],
    });

    fixture = TestBed.createComponent(EventRegistrationFormComponent);
    component = fixture.componentInstance;
  });

  it("should validate required fields", async () => {
    // Given
    component.event = mockEventData;
    fixture.detectChanges();

    // When
    const submitButton = screen.getByRole("button", { name: /register/i });
    fireEvent.click(submitButton);
    await fixture.whenStable();

    // Then
    expect(
      screen.getByText(/personal message is required/i)
    ).toBeInTheDocument();
    expect(
      screen.getByText(/emergency contact is required/i)
    ).toBeInTheDocument();
  });

  it("should submit form with valid data", async () => {
    // Test successful form submission
  });

  it("should handle server validation errors", async () => {
    // Test server-side validation error handling
  });
});
```

---

## 4. End-to-End Testing

### 4.1 Critical User Journeys

**Scope**: Complete user workflows  
**Coverage Target**: 100% of critical paths

#### E2E Test Structure

```typescript
// cypress/e2e/volunteer-registration-flow.cy.ts
describe("Volunteer Registration Flow", () => {
  beforeEach(() => {
    cy.seedDatabase();
    cy.visit("/");
  });

  it("should complete full registration to event participation", () => {
    // Step 1: User Registration
    cy.get("[data-cy=register-button]").click();
    cy.get("[data-cy=email-input]").type("test@example.com");
    cy.get("[data-cy=password-input]").type("SecurePass123!");
    cy.get("[data-cy=name-input]").type("John Volunteer");
    cy.get("[data-cy=submit-registration]").click();

    // Verify email confirmation page
    cy.contains("Check your email for verification");

    // Simulate email verification
    cy.task("verifyEmail", "test@example.com");
    cy.visit("/auth/verify?token=mock-token");

    // Step 2: Profile Setup
    cy.get("[data-cy=complete-profile-button]").click();
    cy.get("[data-cy=location-input]").type("San Francisco, CA");
    cy.get("[data-cy=skill-input]").type("Event Planning");
    cy.get("[data-cy=add-skill-button]").click();
    cy.get("[data-cy=save-profile]").click();

    // Step 3: Find and Register for Event
    cy.get("[data-cy=browse-events]").click();
    cy.get("[data-cy=event-category-filter]").select("Environment");
    cy.get("[data-cy=event-card]").first().click();
    cy.get("[data-cy=register-event-button]").click();

    // Fill registration form
    cy.get("[data-cy=personal-message]").type(
      "Excited to help with this event!"
    );
    cy.get("[data-cy=emergency-contact-name]").type("Jane Doe");
    cy.get("[data-cy=emergency-contact-phone]").type("555-123-4567");
    cy.get("[data-cy=submit-registration]").click();

    // Verify registration success
    cy.contains("Registration confirmed!");
    cy.get("[data-cy=calendar-link]").should("be.visible");

    // Step 4: Verify Dashboard Shows Event
    cy.get("[data-cy=dashboard-link]").click();
    cy.contains("Upcoming Events");
    cy.get("[data-cy=upcoming-event]").should("have.length", 1);
  });

  it("should handle registration conflicts gracefully", () => {
    // Test scheduling conflicts and waitlist scenarios
  });

  it("should allow event cancellation within deadline", () => {
    // Test registration cancellation flow
  });
});
```

#### E2E Testing Checklist

- [ ] **User Registration**: Complete registration and verification flow
- [ ] **Authentication**: Login, logout, password reset flows
- [ ] **Event Discovery**: Search, filter, and browse functionality
- [ ] **Event Registration**: Registration forms and confirmation
- [ ] **Profile Management**: Profile setup and editing
- [ ] **Dashboard Navigation**: All dashboard features work correctly
- [ ] **Mobile Responsiveness**: Critical flows work on mobile devices
- [ ] **Error Scenarios**: Graceful handling of errors and edge cases

### 4.2 Browser and Device Testing

**Scope**: Cross-browser and cross-device compatibility

#### Browser Test Matrix

```typescript
// cypress.config.ts
export default defineConfig({
  e2e: {
    browsers: [
      {
        name: "chrome",
        channel: "stable",
      },
      {
        name: "firefox",
        channel: "stable",
      },
      {
        name: "safari",
        channel: "stable",
      },
      {
        name: "edge",
        channel: "stable",
      },
    ],
    viewports: [
      { width: 375, height: 667 }, // iPhone SE
      { width: 768, height: 1024 }, // iPad
      { width: 1920, height: 1080 }, // Desktop
    ],
  },
});
```

### 4.3 Performance Testing

**Scope**: Core Web Vitals and loading performance

#### Lighthouse CI Configuration

```yaml
# .lighthouserc.yml
ci:
  collect:
    url:
      - http://localhost:4200/
      - http://localhost:4200/events
      - http://localhost:4200/dashboard
    settings:
      chromeFlags: "--no-sandbox --headless"
  assert:
    assertions:
      "categories:performance": ["warn", { minScore: 90 }]
      "categories:accessibility": ["error", { minScore: 95 }]
      "categories:best-practices": ["warn", { minScore: 90 }]
      "categories:seo": ["warn", { minScore: 90 }]
  upload:
    target: "lhci"
    serverBaseUrl: "https://lhci.example.com"
```

---

## 5. Accessibility Testing

### 5.1 Automated Accessibility Testing

**Scope**: WCAG 2.1 AA compliance  
**Coverage Target**: 100% of components

#### axe-core Integration

```typescript
// accessibility.spec.ts
import { axe, toHaveNoViolations } from "jest-axe";

expect.extend(toHaveNoViolations);

describe("Accessibility Tests", () => {
  it("should have no accessibility violations on events page", async () => {
    // Given
    const fixture = TestBed.createComponent(EventsPageComponent);
    fixture.componentInstance.events = mockEventsData;
    fixture.detectChanges();

    // When
    const results = await axe(fixture.nativeElement);

    // Then
    expect(results).toHaveNoViolations();
  });

  it("should be navigable with keyboard only", () => {
    // Test keyboard navigation
    const fixture = TestBed.createComponent(EventCardComponent);
    fixture.detectChanges();

    const element = fixture.nativeElement;
    const firstFocusable = element.querySelector('[tabindex="0"]');

    firstFocusable.focus();
    expect(document.activeElement).toBe(firstFocusable);

    // Simulate tab navigation
    firstFocusable.dispatchEvent(new KeyboardEvent("keydown", { key: "Tab" }));
    // Verify focus moved correctly
  });
});
```

### 5.2 Screen Reader Testing

**Scope**: Screen reader compatibility

#### Screen Reader Test Checklist

- [ ] **Semantic HTML**: Proper heading hierarchy and landmarks
- [ ] **ARIA Labels**: All interactive elements have appropriate labels
- [ ] **Focus Management**: Focus moves logically through the interface
- [ ] **Announcements**: Dynamic content changes are announced
- [ ] **Form Labels**: All form inputs have associated labels
- [ ] **Error Messages**: Errors are announced and properly associated

---

## 6. Visual Regression Testing

### 6.1 Component Visual Tests

**Scope**: UI consistency across changes

#### Percy/Chromatic Integration

```typescript
// visual-regression.spec.ts
describe("Visual Regression Tests", () => {
  it("should match event card appearance", () => {
    const fixture = TestBed.createComponent(EventCardComponent);
    fixture.componentInstance.event = mockEventData;
    fixture.detectChanges();

    // Take screenshot and compare with baseline
    cy.percySnapshot("EventCard - Default State");
  });

  it("should match event card hover state", () => {
    // Test hover states and interactions
  });

  it("should match responsive layouts", () => {
    // Test different viewport sizes
  });
});
```

---

## 7. Security Testing

### 7.1 Frontend Security Tests

**Scope**: Client-side security vulnerabilities

#### Security Test Examples

```typescript
describe("Security Tests", () => {
  it("should sanitize user input to prevent XSS", () => {
    const maliciousInput = '<script>alert("XSS")</script>';
    const component = TestBed.createComponent(UserMessageComponent);
    component.componentInstance.message = maliciousInput;
    fixture.detectChanges();

    const element = fixture.nativeElement;
    expect(element.innerHTML).not.toContain("<script>");
    expect(element.textContent).toContain('alert("XSS")');
  });

  it("should not expose sensitive data in client-side code", () => {
    // Verify no API keys or secrets in bundled code
  });

  it("should validate file uploads on client side", () => {
    // Test file upload restrictions
  });
});
```

---

## 8. API Contract Testing

### 8.1 GraphQL Schema Testing

**Scope**: API contract validation

#### GraphQL Contract Tests

```typescript
describe("GraphQL Contract Tests", () => {
  it("should match expected schema for events query", () => {
    const schema = buildClientSchema(introspectionQuery);
    const query = gql`
      query GetEvents {
        events {
          edges {
            node {
              id
              title
              description
              startTime
              location {
                name
                address
              }
            }
          }
        }
      }
    `;

    const errors = validate(schema, query);
    expect(errors).toHaveLength(0);
  });

  it("should handle schema evolution gracefully", () => {
    // Test backward compatibility
  });
});
```

---

## 9. Test Data Management

### 9.1 Mock Data Strategy

**Scope**: Consistent test data across all test types

#### Mock Data Factory

```typescript
// test-utils/mock-data.ts
export class MockDataFactory {
  static createEvent(overrides?: Partial<Event>): Event {
    return {
      id: "event-1",
      title: "Community Garden Cleanup",
      description: "Help maintain our local community garden",
      status: EventStatus.Published,
      startTime: new Date("2024-03-15T09:00:00Z"),
      endTime: new Date("2024-03-15T12:00:00Z"),
      location: this.createLocation(),
      capacity: this.createCapacity(),
      ...overrides,
    };
  }

  static createUser(overrides?: Partial<User>): User {
    return {
      id: "user-1",
      email: "john.doe@example.com",
      name: "John Doe",
      emailVerified: true,
      ...overrides,
    };
  }

  static createRegistration(overrides?: Partial<Registration>): Registration {
    return {
      id: "reg-1",
      user: this.createUser(),
      event: this.createEvent(),
      status: RegistrationStatus.Confirmed,
      appliedAt: new Date(),
      ...overrides,
    };
  }
}
```

### 9.2 Test Environment Setup

```typescript
// test-setup.ts
import { setupZoneTestEnv } from "jest-preset-angular/setup-env/zone";

setupZoneTestEnv();

// Mock global objects
Object.defineProperty(window, "matchMedia", {
  writable: true,
  value: jest.fn().mockImplementation((query) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: jest.fn(),
    removeListener: jest.fn(),
  })),
});

// Mock IntersectionObserver
global.IntersectionObserver = jest.fn().mockImplementation(() => ({
  observe: jest.fn(),
  unobserve: jest.fn(),
  disconnect: jest.fn(),
}));
```

---

## 10. Continuous Integration Testing

### 10.1 CI/CD Pipeline Testing

**Scope**: Automated testing in CI/CD pipeline

#### GitHub Actions Workflow

```yaml
# .github/workflows/test.yml
name: Test Suite

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        node-version: [18.x, 20.x]

    steps:
      - uses: actions/checkout@v3

      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: ${{ matrix.node-version }}
          cache: "npm"

      - name: Install dependencies
        run: npm ci

      - name: Run linting
        run: npm run lint

      - name: Run unit tests
        run: npm run test:ci

      - name: Run build
        run: npm run build:prod

      - name: Run E2E tests
        run: npm run e2e:ci

      - name: Run Lighthouse CI
        run: npm run lighthouse:ci

      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage/lcov.info

  accessibility:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run Pa11y accessibility tests
        run: npm run test:a11y
```

### 10.2 Quality Gates

- **Code Coverage**: Minimum 90% for unit tests
- **Build Success**: All builds must pass
- **Lint Compliance**: Zero ESLint errors
- **Security Scan**: No high/critical vulnerabilities
- **Performance Budget**: Bundle size within limits
- **Accessibility**: Zero axe-core violations

---

## 11. Test Reporting and Monitoring

### 11.1 Test Reports

```typescript
// jest.config.js
module.exports = {
  coverageReporters: ["html", "lcov", "json-summary"],
  reporters: [
    "default",
    [
      "jest-junit",
      { outputDirectory: "reports", outputName: "jest-junit.xml" },
    ],
    [
      "jest-html-reporters",
      {
        publicPath: "reports",
        filename: "test-report.html",
        expand: true,
      },
    ],
  ],
};
```

### 11.2 Test Metrics Dashboard

- **Test Coverage Trends**: Track coverage over time
- **Test Execution Time**: Monitor test performance
- **Flaky Test Detection**: Identify unreliable tests
- **Browser Compatibility**: Track cross-browser test results
- **Performance Regression**: Monitor Core Web Vitals trends

---

## Testing Implementation Timeline

### Phase 1: Foundation (Weeks 1-2)

- [ ] Set up testing infrastructure and tools
- [ ] Create mock data factories and test utilities
- [ ] Implement basic unit tests for core components
- [ ] Set up CI/CD pipeline with quality gates

### Phase 2: Core Testing (Weeks 3-6)

- [ ] Complete unit tests for all components and services
- [ ] Implement GraphQL integration tests
- [ ] Create E2E tests for critical user journeys
- [ ] Set up accessibility testing automation

### Phase 3: Advanced Testing (Weeks 7-8)

- [ ] Implement visual regression testing
- [ ] Add performance and security tests
- [ ] Create comprehensive test documentation
- [ ] Set up test monitoring and reporting

### Phase 4: Optimization (Weeks 9-10)

- [ ] Optimize test execution performance
- [ ] Implement parallel test execution
- [ ] Create test data management automation
- [ ] Fine-tune quality gates and thresholds

---

## Testing Best Practices Summary

### General Principles

1. **Test Early and Often**: Write tests alongside feature development
2. **Test What Matters**: Focus on user-critical functionality
3. **Maintainable Tests**: Write clear, readable, and maintainable test code
4. **Fast Feedback**: Keep test execution time reasonable
5. **Real User Focus**: Tests should reflect actual user behavior

### Code Quality

- Use descriptive test names that explain the scenario
- Follow AAA pattern (Arrange, Act, Assert)
- Keep tests focused and atomic
- Use data-testid attributes for stable selectors
- Mock external dependencies appropriately

### Automation

- Run tests automatically on every commit
- Include tests in CI/CD pipeline with quality gates
- Generate and monitor test coverage reports
- Set up alerts for test failures and performance regressions

---

**Total Testing Effort Estimate**: 10 weeks parallel with development, ongoing maintenance and updates as features evolve.
