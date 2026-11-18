# Fix HTTP 500 Error When Increasing Product Quantity in Customer Cart — VendorSaathi

You are working on the existing VendorSaathi application. A critical customer-facing bug occurs when a customer clicks the **"+" / Increase Quantity** button for a product already in their cart.

**Observed issue:** The app receives an HTTP `500 Internal Server Error` when attempting to increase the quantity of a cart item.

Your task is to investigate the complete request flow, identify the actual root cause, implement a reliable fix, and verify it with tests.

Do not assume the cause. Inspect the existing implementation before changing anything.

---

## 1. Understand the existing architecture

First inspect:

- `architecture.md`
- Backend source code and cart-related routes, handlers, services, and repositories.
- Flutter customer cart screen and quantity controls.
- API client, request/response models, and error handling.
- PostgreSQL cart tables, constraints, migrations, and queries.
- Existing cart-related tests.

Identify the exact flow:

1. Customer opens the cart.
2. Customer taps the increase quantity button.
3. Flutter sends the API request.
4. Backend authenticates the customer.
5. Backend identifies the cart and cart item.
6. Backend validates the requested quantity and product availability.
7. Backend updates the cart item in PostgreSQL.
8. Backend returns the updated cart.
9. Flutter updates the UI with the new quantity and total price.

Document the endpoint, HTTP method, request payload, handler, service, SQL operation, and response model actually used by this project.

Do not invent a new endpoint if an existing endpoint should be fixed.

## 2. Investigate the HTTP 500 error

Trace the request from Flutter to PostgreSQL.

Inspect these possible failure points:

### Flutter application
- Is the correct cart item ID being sent?
- Is the quantity increment request using the expected payload?
- Is the API URL correct?
- Is the authentication token attached?
- Is the request body serialized correctly?
- Is the app accidentally sending `null`, zero, a negative value, or an incorrect data type?
- Is the request being triggered multiple times by rapid button taps?
- Is the response model correctly handling the backend response?

### Go backend
- Is the route correctly registered?
- Is the authentication middleware working?
- Is the customer ID extracted correctly?
- Is the cart item being retrieved using the correct identifier?
- Are missing cart items handled properly?
- Are database errors being swallowed or incorrectly converted into HTTP 500?
- Are pointer, nil, and type-conversion errors possible?
- Is the handler passing the correct parameters to the service?
- Is the service using the correct transaction boundaries?
- Are errors returned and logged correctly?

### PostgreSQL
Inspect the actual schema and SQL queries.

Check for:
- Incorrect column or table names.
- UUID conversion errors.
- Foreign-key violations.
- Unique constraint violations.
- `NULL` values in non-nullable columns.
- Incorrect quantity types.
- Invalid update queries.
- Transaction failures.
- Deadlocks or concurrent updates.
- Product stock and availability constraints.
- Incorrect joins between carts, cart items, and products.

Do not change the database schema unless the investigation proves that a schema change is necessary.

## 3. Correct quantity update logic

Implement the fix using the existing architecture.

The intended behaviour:

- Clicking `+` increases the selected cart item's quantity by exactly one.
- Clicking `-` decreases it by exactly one.
- Quantity must never become zero or negative unless the existing application explicitly uses zero to remove an item.
- Removing an item must use the existing remove-item functionality.
- Quantity must not exceed available product stock.
- The backend must validate the quantity independently of Flutter.
- The backend must verify that the cart item belongs to the authenticated customer.
- The cart total and item subtotal must be recalculated correctly.
- The product price must come from trusted backend/database data, not from a client-supplied price.
- Updating one item must not accidentally modify another cart item.
- Concurrent quantity updates must not silently corrupt the final quantity.

Use an atomic database update or a transaction with appropriate row locking where required by the existing design.

Preserve existing inventory reservation and checkout behaviour. Do not reserve or deduct stock merely because a customer increases the cart quantity unless the current business logic explicitly requires it.

## 4. Improve backend error handling

Do not return HTTP 500 for expected business errors.

Use appropriate HTTP responses according to the existing API conventions:

| Situation | Expected handling |
|---|---|
| Invalid quantity | 400 Bad Request |
| Unauthenticated request | 401 Unauthorized |
| Customer does not own the cart item | 403 or 404, consistent with existing security policy |
| Cart item does not exist | 404 Not Found |
| Insufficient stock | 409 Conflict |
| Unexpected database/server failure | 500 Internal Server Error |

Return the project's standard API error response format.

Add useful structured server-side logs containing:
- Request/correlation ID, if supported.
- Endpoint and HTTP method.
- Cart ID and cart item ID where safe.
- Authenticated customer ID where safe.
- Operation being performed.
- Actual database or application error.

Never log passwords, JWTs, refresh tokens, or other secrets.

Do not expose SQL statements, stack traces, internal paths, or database credentials to the customer.

## 5. Improve Flutter cart behaviour

Inspect the current cart UI and preserve its existing design.

Implement reliable quantity controls:

- Show a loading state while the quantity update is processing.
- Prevent duplicate requests for the same item while its update is in progress.
- Update the cart only after a successful backend response, or use an optimistic update with reliable rollback on failure.
- On failure, restore the previous quantity and totals.
- Display a useful error message instead of silently failing.
- Refresh or reconcile the cart with the backend when necessary.
- Ensure the displayed item subtotal, cart subtotal, and total are consistent.
- Avoid changing the existing map UI or unrelated screens.

Do not redesign the customer cart screen.

## 6. Add regression tests

### Backend tests

Write tests for:

1. Increasing quantity from 1 to 2.
2. Increasing quantity from 2 to 3.
3. Decreasing quantity.
4. Updating a cart item that does not exist.
5. Updating another customer's cart item.
6. Sending quantity zero.
7. Sending a negative quantity.
8. Sending an invalid quantity type.
9. Increasing quantity beyond available stock.
10. Updating multiple different cart items.
11. Database failure during an update.
12. Concurrent quantity updates.
13. Correct recalculation of item subtotal and cart total.

Use the project's actual PostgreSQL schema and integration-test setup for database-dependent tests. Do not rely exclusively on mocked database tests.

### Flutter tests

Test:
1. Tapping `+` sends the correct request.
2. Successful response updates quantity.
3. Successful response updates totals.
4. Failed request restores the original quantity.
5. Duplicate taps do not cause duplicate updates.
6. Insufficient-stock response displays a useful message.
7. Loading state is displayed correctly.

Run the relevant existing tests as well.

## 7. Verify the complete workflow

Test the complete customer experience:

1. Log in as a customer.
2. Add a product to the cart.
3. Increase its quantity.
4. Confirm the quantity changes.
5. Confirm item subtotal changes.
6. Confirm cart total changes.
7. Decrease quantity.
8. Remove the item.
9. Try increasing quantity beyond available stock.
10. Verify another customer's cart cannot be modified.

Confirm that existing checkout, inventory, and order functionality still works.

## 8. Implementation rules

- Follow the existing project architecture and coding conventions.
- Reuse existing services, repositories, models, and API response structures.
- Avoid unnecessary dependencies.
- Avoid duplicate business logic.
- Do not hardcode product IDs, cart IDs, customer IDs, or stock values.
- Do not bypass authentication or authorization.
- Do not disable validation to hide the error.
- Do not swallow errors or return fake success responses.
- Do not modify production data.
- Do not run destructive database commands.
- Do not change unrelated features.
- Do not deploy, push, or publish changes without explicit approval.

## 9. Final report

Create `CART_QUANTITY_500_FIX_REPORT.md` containing:

1. Root cause of the HTTP 500.
2. Exact endpoint and request payload involved.
3. Backend files modified.
4. Flutter files modified.
5. Database queries or schema changes, if any.
6. Explanation of the implemented fix.
7. Tests added.
8. Test commands and actual results.
9. Remaining issues or limitations.
10. Manual verification instructions.

**Important:** Do not declare this bug fixed just because the application builds. The actual quantity-increase API request must succeed against the database, and regression tests must pass.

Start by investigating and reproducing the failure. Then fix the root cause and verify the entire cart quantity workflow.