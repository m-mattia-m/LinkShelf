// @testing-library/jest-dom@7's own `vitest` module augmentation declares
// `Assertion<T = any>` (one generic param), but vitest 5 changed the real
// `Assertion` interface to `Assertion<R, T>` (two required params, no
// defaults) - the parameter count mismatch means TypeScript's declaration
// merging silently fails to attach jest-dom's matchers (toBeInTheDocument,
// toHaveValue, etc.) to vitest's `expect(...)` return type. This re-declares
// the merge with the correct shape until jest-dom ships its own vitest 5
// support. Safe to delete once that lands upstream.
import type { TestingLibraryMatchers } from '@testing-library/jest-dom/matchers'

/* eslint-disable @typescript-eslint/no-explicit-any, @typescript-eslint/no-empty-object-type -- mirrors jest-dom's own upstream augmentation shape exactly, generic defaults included */
declare module 'vitest' {
  interface Assertion<R = any, T = any> extends TestingLibraryMatchers<T, R> {}
  interface AsymmetricMatchersContaining extends TestingLibraryMatchers<any, any> {}
}
