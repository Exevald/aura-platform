# Aura frontend

React, TypeScript and Vite client organized with Model–View–Presenter:

- `src/model` contains API contracts and transport code and does not depend on React;
- `src/views` contains passive UI components that render props and emit user events;
- `src/presenters` coordinates requests, state, navigation, autosave and error handling;
- `src/App.tsx` declares routes and connects them to presenters.

Run `npm run lint`, `npm run test` and `npm run build` before delivery.

## Session behaviour

The browser keeps only the server-issued HttpOnly cookie; session identifiers are
never exposed to JavaScript. Authentication changes are propagated immediately
between tabs of the same origin with `BroadcastChannel`. When a tab becomes
visible or receives focus, it calls `/api/bff/v1/me` again, which also picks up
login and logout performed in another product on a sibling subdomain.

The local product addresses (`aura-platform.documents`, `aura-platform.calendar`,
and others) use a central login at `aura-platform.auth`. The auth service issues
a one-time code, then the destination receives its own HttpOnly cookie. Local
HTTP uses host-only cookies with `AUTH_COOKIE_DOMAIN` empty. Production deployments
need HTTPS, `AUTH_COOKIE_SECURE=true`, and a deployment-specific SSO host allowlist.
