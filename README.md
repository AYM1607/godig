# Godig
A simple tunneling service to expose local services publicly.

## Architecture

```
┌────────────┐       ┌────────────┐       ┌────────────┐
│            │Bearer │            │API Key│            │
│   Client   │──────▶│   Server   │◀─────▶│  Service   │
│            │       │            │       │            │
└────────────┘       └────────────┘       └────────────┘
     │                      │                      │
     │                      │                      │
     ▼                      ▼                      ▼
Bearer Auth            Domain Routing     Exposes local http
```

## Core features
- L7 tunneling
- Domain based routing
- Service requests id which is also their subdomain
- API key (pre-shared) based auth between Server and Service
- Bearer authorization between Clients and Server
- Service generates bearer tokens on initial connection
- SSE streaming support

## Philosophy
**No scope creep**, new features will most likely not be added.
Conceived for developers, designed and created for production.
