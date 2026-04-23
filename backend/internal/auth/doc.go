// Package auth will contain Clerk token validation and auth-related middleware.
//
// Keeping auth isolated from route handlers helps keep permission checks
// consistent and avoids scattering JWT logic across the codebase.
package auth
