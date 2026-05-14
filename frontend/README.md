# Frontend

This is the Vue 3 + Vite frontend for the Masters Pool rewrite.

## Auth setup

The frontend now supports two modes:

- Clerk mode when `VITE_CLERK_PUBLISHABLE_KEY` is present
- mock-compatible mode when that key is missing

In Clerk mode, the Vue app:

- initializes `@clerk/vue` in `src/main.js`
- shows the dedicated `/sign-in` route
- uses Clerk browser session state to guard `/enter` and `/admin`
- reads a session token from Clerk and attaches it as a bearer token on protected API calls

The backend then verifies that JWT against Clerk JWKS. This frontend does not use any Next.js-specific Clerk setup.

Create a frontend env file before running the real Clerk browser flow:

```bash
cp .env.example .env
```

Then set:

```bash
VITE_CLERK_PUBLISHABLE_KEY=pk_test_replace_me
```

The repo root `.env` must also contain the matching backend Clerk settings and should disable mock auth for the real proof:

```bash
MOCK_AUTH_ENABLED=false
CLERK_SECRET_KEY=sk_test_replace_me
CLERK_JWKS_URL=https://your-clerk-domain/.well-known/jwks.json
CLERK_ISSUER=https://your-clerk-domain
CLERK_AUTHORIZED_PARTIES=http://localhost:5173,http://127.0.0.1:5173
```

The recommended Clerk session token claims for this app are:

```json
{
  "email": "{{user.primary_email_address}}",
  "name": "{{user.first_name}} {{user.last_name}}",
  "role": "{{user.public_metadata.role}}"
}
```

Without that key, the app still boots so backend and UI work can continue against the temporary mock-auth path.

## Scripts

```bash
npm install
npm run dev
npm run build
```
