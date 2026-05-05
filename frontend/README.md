# Frontend

This is the Vue 3 + Vite frontend for the Masters Pool rewrite.

## Auth setup

The frontend now supports two modes:

- Clerk mode when `VITE_CLERK_PUBLISHABLE_KEY` is present
- mock-compatible mode when that key is missing

Create a frontend env file before running the real Clerk browser flow:

```bash
cp .env.example .env
```

Then set:

```bash
VITE_CLERK_PUBLISHABLE_KEY=pk_test_replace_me
```

Without that key, the app still boots so backend and UI work can continue against the temporary mock-auth path.

## Scripts

```bash
npm install
npm run dev
npm run build
```
