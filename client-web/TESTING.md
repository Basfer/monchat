# Frontend Testing Guide

This guide explains how to run tests for the Matrix Messenger client-web application.

## Prerequisites

Make sure you have all dependencies installed:

```bash
npm install
```

## Running Tests

### Run all tests

```bash
npm test
```

### Run tests in watch mode (for development)

```bash
npm test -- --watch
```

### Run tests and generate coverage report

```bash
npm test -- --coverage
```

### Run a specific test file

```bash
npm test -- UserSearchModal.test.tsx
```

### Run tests matching a pattern

```bash
npm test -- --testNamePattern="renders modal when open"
```

## Test Structure

Tests are located in `__tests__` directories next to the corresponding source files:

```
client-web/src/
├── components/
│   ├── __tests__/
│   │   └── UserSearchModal.test.tsx    # UserSearchModal component tests
│   ├── UserSearchModal.tsx
│   └── ...
└── ...
```

## Current Tests

### UserSearchModal (`UserSearchModal.test.tsx`)

Tests for the [`UserSearchModal`](src/components/UserSearchModal.tsx) component:

| Test Suite | Description |
|------------|-------------|
| **Rendering** | Tests modal open/close states, auto-focus behavior |
| **Close Handlers** | Tests Escape key, overlay click, content click |
| **Search Functionality** | Tests search hint, debounce, loading state, empty results |
| **Search Results Display** | Tests user info display, avatars, message buttons |
| **Chat Creation** | Tests button states, success flow, error handling |
| **Modal Reset on Close** | Tests clearing of search query and results |

## Technology Stack

- **Jest** - Test runner and assertion library
- **ts-jest** - TypeScript support for Jest
- **Testing Library** - React component testing utilities
- **jsdom** - DOM environment for Node.js

## Configuration

Jest configuration is located in [`jest.config.js`](jest.config.js):

- Test files: `**/__tests__/**/*.test.tsx`, `**/__tests__/**/*.test.ts`
- Environment: jsdom
- Setup: `@testing-library/jest-dom`
- Coverage output: `<rootDir>/coverage`

## Mocking

The project uses Jest's built-in mocking for:

- **Zustand stores** - Mocked via `jest.mock()` at module level
- **API services** - Mocked via `jest.mock()` at module level

Example mock pattern used in the project:

```typescript
jest.mock('../../store/authStore', () => ({
  useAuthStore: jest.fn(),
}));
```

## Adding New Tests

1. Create a new `.test.tsx` or `.test.ts` file in a `__tests__` directory
2. Use Jest's `describe` and `test` functions
3. Use Testing Library functions (`render`, `screen`, `fireEvent`, etc.)
4. Mock external dependencies (stores, APIs) using `jest.mock()`
