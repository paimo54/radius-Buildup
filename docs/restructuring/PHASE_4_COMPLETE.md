# Phase 4 Complete — API Route Thinning

## Summary

Phase 4 focused on making route handlers thin (parse → validate → call service → respond), extracting business logic into service files, and standardising HTTP responses with shared helpers. A critical connection-pool bug was also fixed across the entire API surface.

---

## What Was Done

### 1. API Response Helpers (`src/lib/api-response.ts`)

Created centralised response helpers to replace ad-hoc `NextResponse.json(…, { status: N })` calls:

| Helper | Status | Usage |
|--------|--------|-------|
| `ok(data, meta?)` | 200 | Successful GET/PUT/PATCH |
| `created(data)` | 201 | Successful POST create |
| `badRequest(error)` | 400 | Validation failures |
| `unauthorized()` | 401 | Missing session |
| `forbidden()` | 403 | Insufficient role |
| `notFound(resource?)` | 404 | Entity not found |
| `conflict(error)` | 409 | Duplicate / constraint violation |
| `serverError(error?)` | 500 | Unexpected errors |
| `validationError(errors)` | 422 | Zod validation failures |

### 2. Parse Helpers (`src/lib/parse-body.ts`)

Zod-powered request helpers:
- `parseBody<T>(request, schema)` — parse and validate JSON body
- `parseQuery<T>(searchParams, schema)` — parse and validate URL query params
- `requireParam(value, name)` — reject with 400 if param missing

### 3. New Service Files

| File | Extracted From | Exports |
|------|---------------|---------|
| `src/server/services/pppoe.service.ts` | `pppoe/users/route.ts` (746 lines) | `listPppoeUsers`, `getPppoeUserById`, `createPppoeUser`, `updatePppoeUser`, `deletePppoeUser` |
| `src/server/services/hotspot.service.ts` | `hotspot/voucher/route.ts` (604 lines) | `CODE_TYPES`, `listVouchers`, `generateVouchers`, `deleteVouchers`, `patchVouchers` |

### 4. Thinned Route Files

| Route | Before | After | Reduction |
|-------|--------|-------|-----------|
| `pppoe/users/route.ts` | 746 lines | 89 lines | −88% |
| `pppoe/users/[id]/route.ts` | ~40 lines | 18 lines | −55% |
| `hotspot/voucher/route.ts` | 604 lines | 111 lines | −82% |
| `invoices/route.ts` | 618 lines | ~580 lines | helpers applied |
| `notifications/route.ts` | 155 lines | ~80 lines | helpers applied |
| `cron/route.ts` | — | — | auth helpers |
| `cron/status/route.ts` | — | — | auth helpers |
| `cron/telegram/route.ts` | — | — | auth/error helpers |

### 5. Critical Bug Fix — Connection Pool Exhaustion

**Problem**: 25 route files were calling `new PrismaClient()` at module level. Each hot-reload or serverless cold start created a new connection pool, rapidly exhausting available MySQL connections.

**Fix**: Replaced all instances with the shared singleton from `@/lib/prisma`:

```typescript
// BEFORE (each file)
import { PrismaClient } from '@prisma/client';
const prisma = new PrismaClient();

// AFTER
import { prisma } from '@/lib/prisma';
```

**Files fixed** (25 total):

| Directory | Files |
|-----------|-------|
| `admin/` | `laporan/route.ts`, `invoices/import/route.ts` |
| `agent/` | `record-sales/route.ts`, `login/route.ts`, `dashboard/route.ts` |
| `company/` | `route.ts` |
| `cron/` | `status/route.ts` |
| `evoucher/` | `order/[token]/route.ts` |
| `hotspot/agents/` | `route.ts`, `balance/route.ts`, `[id]/history/route.ts` |
| `invoices/` | `route.ts`, `generate/route.ts`, `check/route.ts`, `by-token/[token]/route.ts` |
| `keuangan/` | `transactions/route.ts`, `export/route.ts`, `categories/route.ts` |
| `network/routers/` | `route.ts`, `status/route.ts`, `[id]/uplinks/route.ts`, `[id]/interfaces/route.ts`, `[id]/ping-olt/route.ts`, `[id]/detect-public-ip/route.ts`, `[id]/setup-radius/route.ts`, `[id]/setup-isolir/route.ts` |

### 6. Protected Routes (Not Modified)

The following routes were intentionally left unchanged to avoid breaking production integrations:

- `radius/authorize`, `radius/accounting`, `radius/post-auth`, `radius/coa` — FreeRADIUS protocol endpoints with custom response formats
- All routes' business logic and URL paths — only response helpers and imports were updated

---

## TypeScript Status

```
npx tsc --noEmit  →  exit code 0 (no errors)
```

---

## Pattern Reference

### Thin Route Handler Pattern

```typescript
export async function GET(request: NextRequest) {
  const session = await getServerSession(authOptions);
  if (!session) return unauthorized();

  const { searchParams } = new URL(request.url);
  try {
    const result = await myService.listItems({ filter: searchParams.get('filter') });
    return ok({ items: result });
  } catch {
    return serverError();
  }
}
```

### Service Error Convention

Services throw errors with a `.code` property for route handlers to inspect:

```typescript
// In service
const err = new Error('Username already taken');
(err as any).code = 'DUPLICATE_USERNAME';
throw err;

// In route handler
} catch (error: any) {
  if (error.code === 'DUPLICATE_USERNAME') return conflict(error.message);
  return serverError();
}
```
