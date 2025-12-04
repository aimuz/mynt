# Mynt NAS - Claude AI Context

This document provides Claude AI with essential context about the Mynt NAS project to enable effective assistance with development tasks.

## Important Development Notes

> [!CAUTION]
> **Critical Requirements for All Development Work**

1. **Be humble & honest** - NEVER overstate what you got done or what actually works in commits, PRs or in messages to the user.
2. **All changes must be tested** - if you're not testing your changes, you're not done.
3. **Get your tests to pass**. If you didn't run the tests, your code does not work.
4. **Follow existing code style** - check neighboring files for patterns
5. **Use bun, not npm** - This project uses bun as the package manager for frontend

## Project Overview

**Mynt** is a modern, enterprise-grade Network Attached Storage (NAS) system built with Go and Svelte. The project aims to deliver:
- **Enterprise-grade stability**: Robust ZFS integration with comprehensive error handling
- **iOS-like usability**: Glassmorphism UI with smooth animations and intuitive interactions
- **Modern architecture**: Clean separation between backend (Go) and frontend (Svelte 5)

**Module Path**: `go.aimuz.me/mynt`

## Architecture

### Technology Stack

#### Backend (Go 1.25.4)
- **Web Framework**: Standard library `net/http`
- **Storage**: ZFS management via system commands
- **Database**: SQLite (modernc.org/sqlite)
- **Authentication**: JWT (golang-jwt/jwt/v5)
- **Migrations**: Goose (pressly/goose/v3)

#### Frontend (Svelte 5)
- **Framework**: SvelteKit 2
- **Styling**: Tailwind CSS 4 (with glassmorphism design)
- **Build Tool**: Vite 7
- **Package Manager**: Bun (NOT npm)
- **Icons**: Lucide Svelte
    - **Package name**: Always import from `@lucide/svelte` (the correct package name)
    - ⚠️ Do NOT use `lucide-svelte` (deprecated/incorrect package name)
- **Charts**: Chart.js
- **Deployment**: Static adapter (embedded in Go binary)

### Project Structure

```
mynt/
├── cmd/                    # Entry points
│   ├── mynt/              # CLI commands
│   └── myntd/             # Main daemon
├── internal/              # Private application code
│   └── api/               # HTTP API server and handlers
├── auth/                  # Authentication and JWT logic
├── disk/                  # Disk management
├── event/                 # Event bus for notifications
├── monitor/               # System monitoring
├── share/                 # SMB/NFS share management
├── store/                 # Database repositories
├── sysexec/               # System command execution abstraction
├── task/                  # Background task management
├── user/                  # User management
├── zfs/                   # ZFS operations wrapper
├── web-ui/                # Svelte frontend
│   └── src/
│       ├── routes/        # SvelteKit routes
│       │   ├── desktop/   # Main desktop UI
│       │   ├── login/     # Login page
│       │   └── setup/     # Initial setup wizard
│       └── lib/
│           ├── api.ts     # API client
│           ├── apps/      # Desktop applications
│           ├── components/ # Reusable UI components
│           ├── widgets/   # Desktop widgets
│           └── stores/    # Svelte stores
└── tests/                 # Integration tests
```

## Key Design Patterns

### Backend Patterns

#### 1. Dependency Injection
Services accept dependencies through constructors:
```go
type Manager struct {
    exec sysexec.Executor
    bus  *event.Bus
}

func NewManager(exec sysexec.Executor, bus *event.Bus) *Manager {
    return &Manager{exec: exec, bus: bus}
}
```

#### 2. System Command Abstraction
All system commands use the `sysexec.Executor` interface for testability:
- Production: `sysexec.NewExecutor()` - executes real commands
- Testing: `sysexec.NewMock()` - simulates command execution

See [sysexec/README.md](./sysexec/README.md) for details.

#### 3. Event-Driven Architecture
Components communicate via the event bus:
```go
bus.Publish(event.Event{
    Type: event.TypePoolCreated,
    Data: poolData,
})
```

#### 4. Repository Pattern
Database access is abstracted through repositories in `store/`:
- `ConfigRepo` - System configuration
- `NotificationRepo` - Notification storage

#### 5. Response Helpers
API responses use consistent helper functions from `internal/api/response.go`:
- `WriteJSON()` - Success responses
- `WriteError()` - Error responses

### Frontend Patterns

#### 1. SvelteKit Routing
- `/` - Landing page
- `/login` - Authentication
- `/setup` - Initial configuration wizard
- `/desktop` - Main application (desktop environment)

#### 2. Desktop UI Architecture
The desktop uses a window manager pattern:
- **Desktop component**: Manages windows and provides context
- **Apps**: Full-screen applications (StorageApp, ShareManagementApp, etc.)
- **Windows**: Draggable, resizable modals for actions
- **Widgets**: Dashboard elements (ClockWidget, SystemStatusWidget, etc.)

#### 3. Context API
Child components access desktop functionality via Svelte context:
```typescript
const { openWindow, closeWindow } = getContext('desktop');
```

#### 4. API Client
Centralized API communication in `lib/api.ts`:
```typescript
export const api = {
    listPools: () => fetch('/api/pools').then(r => r.json()),
    createPool: (data) => fetch('/api/pools', { method: 'POST', body: JSON.stringify(data) })
    // ...
};
```

#### 5. Error Handling Strategy
- **Svelte boundaries**: `<svelte:boundary>` for synchronous errors
- **Defensive programming**: Try-catch blocks for async operations
- **Null safety**: Always check for null API responses
- **Graceful degradation**: Individual component failures don't freeze the UI

## Development Workflow

### Building

```bash
# Build backend
make build                    # Output: bin/myntd

# Build frontend (done automatically during Go build)
cd web-ui
bun run build                 # Output: build/

# Run development server
make run                      # Starts myntd

# Frontend dev mode (with hot reload)
cd web-ui
bun run dev
```

### Testing

```bash
# Backend tests
make test                     # Unit tests
make test-integration         # Integration tests
make coverage                # Generate coverage report

# Frontend
cd web-ui
bun run check                # Type checking and Svelte validation
```

### Frontend Development Server
When developing the UI:
1. Run backend: `make run` (serves API on port 8080)
2. Run frontend: `cd web-ui && bun run dev` (serves UI on port 5173)
3. The frontend proxies API requests to the backend

## API Endpoints

All API endpoints are defined in [internal/api/server.go](./internal/api/server.go). The API follows RESTful conventions and is organized into the following groups:

- **Authentication**: Setup, login, JWT-based auth
- **Storage Management**: Disks, ZFS pools, datasets
- **Shares**: SMB/NFS share configuration
- **Users**: User account management
- **Notifications**: Notification system with SSE support

To explore available endpoints, check the `routes()` method in `server.go`, which maps all HTTP routes to their handlers. Each handler method includes documentation about its purpose, request format, and response structure.

## Code Style Guidelines

### Go Code
- Use standard Go formatting (`gofmt`)
- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Error messages: lowercase, no trailing punctuation
- Package comments: document purpose and usage
- Exported functions: always include godoc comments

### TypeScript/Svelte
- Use TypeScript for type safety
- Follow Svelte 5 runes syntax (`$state`, `$derived`, `$effect`)
- Component files: PascalCase (e.g., `StorageApp.svelte`)
- Use Tailwind CSS classes for styling
- Maintain glassmorphism aesthetic: translucent backgrounds, blur effects, smooth animations

### Design Principles
1. **Glassmorphism UI**: Translucent panels with backdrop-blur
2. **Smooth animations**: Use CSS transitions and animations
3. **Dark mode**: Default color scheme
4. **Responsive design**: Support various screen sizes
5. **Accessibility**: Proper labels, ARIA attributes, keyboard navigation

## Common Tasks

### Adding a New API Endpoint
1. Add handler method to `internal/api/server.go`
2. Register route in `routes()` method
3. Add error handling with `WriteError()` and `WriteJSON()`
4. Update frontend `lib/api.ts` with corresponding function
5. Test the endpoint with `make test`

### Adding a New Desktop App
1. Create Svelte component in `web-ui/src/lib/apps/`
2. Import in desktop page `web-ui/src/routes/desktop/+page.svelte`
3. Add icon to dock
4. Use `getContext('desktop')` for window operations
5. Test with `bun run check` and manual testing

### Adding a New Widget
1. Create Svelte component in `web-ui/src/lib/widgets/`
2. Add to desktop layout
3. Implement error boundary for robustness
4. Test error handling scenarios

### Database Migration
1. Create new file in `store/migrations/` following sequence
2. Use Goose syntax (SQL with special comments)
3. Test with `make test`

## Testing Philosophy

### Backend Testing
- **Unit tests**: Mock external dependencies using `sysexec.NewMock()`
- **Integration tests**: Test against real SQLite database
- **No root required**: Tests should not require privileged access
- **Fast feedback**: Tests should complete quickly
- **All tests must pass**: Don't commit broken tests

### Frontend Testing
- **Type safety**: Use TypeScript and `bun run check`
- **Manual testing**: Use browser dev tools
- **Error testing**: Temporarily inject errors to verify error boundaries
- **Svelte validation**: Run `bun run check` before committing

## Known Limitations & Gotchas

1. **ZFS requires root**: Production deployment needs root or appropriate capabilities
2. **SQLite in-memory**: Tests use in-memory databases
3. **API null responses**: Always check for `null` in array responses - API may return null instead of empty arrays
4. **Svelte 5 boundaries**: Only catch synchronous errors; use try-catch for async operations
5. **Static build**: Web UI is embedded in binary; rebuild Go binary after UI changes
6. **Bun only**: Always use `bun`, never `npm` for frontend operations

## Debugging Tips

### Backend
- Check logs for error messages
- Use `go test -v` for verbose test output
- Inspect database: `sqlite3 mynt.db`
- Check ZFS command execution with mock executor in tests

### Frontend
- Browser DevTools Console for JavaScript errors
- Network tab for API request/response inspection
- Svelte DevTools extension for component inspection
- Check for null API responses causing runtime errors

## Security Considerations

- JWT authentication for all protected endpoints
- Admin-only operations require admin role
- Password hashing with bcrypt
- HTTPS recommended for production
- Validate all user input on backend
- Never trust client-side validation alone

## Performance Optimization

### Backend
- Connection pooling for database
- Efficient ZFS command caching where appropriate
- Async event publishing to avoid blocking
- Minimize system command execution overhead

### Frontend
- Static site generation where possible
- Lazy loading for heavy components
- Debounce rapid API calls
- Optimize animations (use `transform` and `opacity`)
- Minimize re-renders with proper Svelte reactivity

## Contributing Guidelines

When making changes:
1. **Follow existing code patterns** - check neighboring files for reference
2. **Add tests for new functionality** - untested code doesn't work
3. **Ensure all tests pass** - run `make test` for backend, `bun run check` for frontend
4. **Update documentation** - keep this file and code comments current
5. **Maintain glassmorphism design aesthetic** - consistent UI/UX across the app
6. **Consider error handling and edge cases** - especially null API responses
7. **Be honest about what works** - never overstate functionality

## Additional Resources

- [Svelte 5 Documentation](https://svelte.dev/docs/svelte/overview)
- [SvelteKit Documentation](https://svelte.dev/docs/kit/introduction)
- [ZFS Documentation](https://openzfs.github.io/openzfs-docs/)
- [Go Standard Library](https://pkg.go.dev/std)
- [Bun Documentation](https://bun.sh/docs)
