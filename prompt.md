# VendorSaathi — Build the Complete Application

You are the lead software architect and senior full-stack engineer responsible for building VendorSaathi.

The attached `architecture.md` contains the complete technical architecture, technology stack, database design, API specifications, business logic, coding standards, security requirements, and implementation guidelines.

Your responsibility is to turn this architecture into a fully functional, production-ready application.

## 1. Primary Instruction

Read and understand the entire `architecture.md` before writing code.

**Start implementing the application immediately. Do not spend time creating architecture reviews, audit reports, or separate implementation plans.**

Treat `architecture.md` as the primary source of truth throughout development.

Do not redesign the architecture or unnecessarily change the selected technologies.

## 2. Technology Stack

Follow the exact stack specified in `architecture.md`, including:

- Backend: Go with Gin.
- Database: PostgreSQL.
- Mobile application: Flutter.
- Admin dashboard: React.
- State management and supporting libraries as specified in the architecture.
- Offline storage and synchronization: SQLite.
- Authentication: JWT.
- Notifications: Firebase Cloud Messaging.
- External integrations: Google Maps and OpenWeather where specified.
- Deployment: Docker and the documented infrastructure.

Do not replace these technologies with alternatives without a compelling technical reason.

## 3. Development Approach

Build the application incrementally in dependency order.

### Phase 1 — Backend Foundation
- Initialize the Go backend.
- Establish the documented project structure.
- Implement configuration and environment management.
- Implement PostgreSQL connectivity.
- Create database migrations.
- Implement dependency injection.
- Set up Gin routing and middleware.
- Implement authentication and authorization.
- Implement structured logging and error handling.

### Phase 2 — Core Backend Modules
Implement the modules specified in `architecture.md`, including:

- User and role management.
- Vendor management.
- Customer management.
- Product management.
- Inventory management.
- Categories.
- Location and nearby vendor discovery.
- Cart management.
- Order management.
- Order fulfilment.
- Payment records.
- Split payments.
- Udhaar ledger.
- Notifications.
- Sales summaries and analytics.
- Admin management and audit logs.

Implement actual database operations and business logic. Do not substitute mock responses for real functionality.

### Phase 3 — Flutter Application
Build the Vendor and Customer experiences within the Flutter application.

Implement:
- Authentication.
- Role-based navigation.
- Vendor dashboard.
- Product and inventory management.
- Customer home screen.
- Nearby vendor discovery.
- Product browsing.
- Cart and checkout.
- Order tracking.
- Payment records.
- Udhaar functionality.
- Notifications.
- Offline caching and synchronization as specified.

Maintain a clean, consistent, intuitive UI.

### Phase 4 — React Admin Dashboard
Build the administrative dashboard.

Implement:
- Admin authentication.
- Dashboard overview.
- Vendor management.
- Customer management.
- Product oversight.
- Order monitoring.
- Payment monitoring.
- Udhaar oversight where authorized.
- Analytics and reports.
- Audit logs.
- Role-based access control.

### Phase 5 — Integration
Connect the Flutter application and React dashboard to the actual Go backend.

Verify:
- API contracts.
- Authentication.
- Authorization.
- Data consistency.
- Order workflows.
- Inventory updates.
- Payment recording.
- Notifications.
- Error handling.

### Phase 6 — Testing and Production Readiness
- Write unit tests.
- Write integration tests.
- Test critical business workflows.
- Test concurrent inventory operations.
- Test authentication and authorization.
- Test offline synchronization.
- Fix discovered bugs.
- Verify database migrations.
- Configure Docker.
- Prepare deployment documentation.

## 4. Non-Negotiable Engineering Rules

1. Follow `architecture.md` consistently.
2. Use clean architecture and modular design as documented.
3. Use dependency injection.
4. Avoid unnecessary global mutable variables.
5. Never bypass backend authorization.
6. Enforce vendor and customer data ownership.
7. Use database transactions for atomic operations.
8. Implement idempotency where required.
9. Follow the documented order state machine.
10. Maintain financial accuracy using integer paise for INR.
11. Keep the Udhaar ledger append-only.
12. Implement proper offline synchronization and conflict handling.
13. Never hardcode credentials, secrets, or API keys.
14. Never commit `.env` files.
15. Write maintainable, reusable, documented code.
16. Keep frontend and backend API contracts consistent.
17. Do not introduce unnecessary dependencies.
18. Do not leave critical functionality as TODOs or placeholders.

## 5. How You Must Work

- Start with the backend foundation and proceed in dependency order.
- Implement complete features rather than generating large amounts of disconnected boilerplate.
- Before implementing a module, refer to its corresponding architecture section.
- After implementing each module, run the relevant tests and fix errors.
- Keep the application runnable throughout development.
- Maintain accurate documentation as the implementation evolves.
- Track completed and pending work in `DEVELOPMENT_PROGRESS.md`.
- Do not repeatedly ask me for permission to proceed to the next development phase.

If you encounter an unresolved business decision in `architecture.md`, use a safe, reversible implementation consistent with the documented assumptions and record the decision in `DEVELOPMENT_PROGRESS.md`. Ask me only when the decision genuinely blocks implementation or could cause irreversible consequences.

If you encounter technical problems, investigate and resolve them instead of abandoning the task.

## 6. Execution Priority

Prioritize in this order:

1. Correct backend architecture and database.
2. Working APIs and business logic.
3. Flutter application.
4. React admin dashboard.
5. Complete integration.
6. Testing, optimization, and deployment.

Do not sacrifice correctness or maintainability for speed.

## 7. Start Now

Begin by reading `architecture.md`, inspecting the existing project structure as part of normal development, and immediately implementing the first development phase.

Do not generate an architecture review.
Do not generate a preliminary audit.
Do not stop after planning.
Do not wait for my approval between phases.

**Your objective is to BUILD VendorSaathi into a complete, integrated, functional application according to `architecture.md`. Start coding now.**