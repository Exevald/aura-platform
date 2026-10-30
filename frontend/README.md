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

For production products hosted on sibling subdomains, configure the auth service
with their shared parent domain in `AUTH_COOKIE_DOMAIN` and use HTTPS everywhere.
For local development leave the domain empty and consistently use `localhost`
(do not mix it with `127.0.0.1`). A session created with `rememberMe: true`
survives a browser restart; without it the browser cookie is session-only.
