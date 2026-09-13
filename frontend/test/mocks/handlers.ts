import { http, HttpResponse } from 'msw'
import type { ErrorDetail } from '~~/api'
import {
  LinkOrderResponseBodyToJSON,
  LinkToJSON,
  SectionOrderResponseBodyToJSON,
  SectionToJSON,
  SettingBatchResponseBodyToJSON,
  SettingPageBodyToJSON,
  ShelfToJSON,
  StatisticToJSON,
  ThemeGroupedResponseBodyToJSON,
  ThemeToJSON,
  TokenPairToJSON,
  UserToJSON
} from '~~/api'
import {
  buildSettingPageBody,
  buildStatistic,
  buildThemeGrouped,
  buildTokenPair,
  buildUser
} from './factories'

// Matches the backend's huma-style error body (see useApiError.ts). This one
// is intentionally NOT run through a generated *ToJSON() converter - the
// error shape is snake_case-free and consistent, unlike the success models.
export function errorResponse(status: number, detail: string, errors: ErrorDetail[] = []) {
  return HttpResponse.json({ status, detail, errors }, { status })
}

const BASE = 'http://localhost:8085'

// The generated OpenAPI models are camelCase in TypeScript but several of
// them serialize to snake_case (or otherwise-renamed) JSON on the wire -
// see api/models/*.ts's own *FromJSON()/*ToJSON() pairs. Every handler here
// runs its body through the matching *ToJSON() converter so responses match
// exactly what the real backend sends, instead of hand-written JSON that
// silently drifts from the wire format.
export const handlers = [
  // --- Auth ---
  http.post(`${BASE}/v1/auth/login`, () => HttpResponse.json(TokenPairToJSON(buildTokenPair()))),
  http.post(`${BASE}/v1/auth/logout`, () => new HttpResponse(null, { status: 204 })),
  http.post(`${BASE}/v1/auth/refresh`, () => HttpResponse.json(TokenPairToJSON(buildTokenPair()))),
  http.post(`${BASE}/v1/auth/resend-verification`, () => new HttpResponse(null, { status: 204 })),
  http.post(`${BASE}/v1/auth/set-password`, () => new HttpResponse(null, { status: 204 })),
  http.post(`${BASE}/v1/auth/verify-email`, () => new HttpResponse(null, { status: 204 })),
  http.get(`${BASE}/v1/auth/oidc/login`, () => HttpResponse.json({
    authorization_url: 'https://idp.example.com/authorize?state=abc',
    state: 'abc'
  })),
  http.post(`${BASE}/v1/auth/oidc/callback`, () => HttpResponse.json(TokenPairToJSON(buildTokenPair()))),

  // --- Users ---
  http.get(`${BASE}/v1/users/me`, () => HttpResponse.json(UserToJSON(buildUser({ id: 'user-1' })))),
  http.get(`${BASE}/v1/users`, () => HttpResponse.json([buildUser({ id: 'user-1' })].map(UserToJSON))),
  http.get(`${BASE}/v1/users/:userId`, ({ params }) => HttpResponse.json(UserToJSON(buildUser({ id: params.userId as string })))),
  http.post(`${BASE}/v1/users`, () => HttpResponse.json(UserToJSON(buildUser()))),
  http.put(`${BASE}/v1/users/:userId`, ({ params }) => HttpResponse.json(UserToJSON(buildUser({ id: params.userId as string })))),
  http.patch(`${BASE}/v1/users/:userId/password`, () => new HttpResponse(null, { status: 204 })),
  http.patch(`${BASE}/v1/users/:userId/verify`, ({ params }) => HttpResponse.json(UserToJSON(buildUser({ id: params.userId as string, emailVerified: true })))),
  http.delete(`${BASE}/v1/users/:userId`, () => new HttpResponse(null, { status: 204 })),

  // --- Shelves ---
  http.get(`${BASE}/v1/shelves`, () => HttpResponse.json([])),
  http.get(`${BASE}/v1/shelves/by-path/:path`, () => HttpResponse.json({
    id: 'shelf-1',
    path: 'shelf-1',
    description: '',
    sections: []
  })),
  http.get(`${BASE}/v1/shelves/:shelfId`, ({ params }) => HttpResponse.json(ShelfToJSON({
    id: params.shelfId as string,
    path: `shelf-${params.shelfId}`,
    description: '',
    domain: '',
    icon: 'i-lucide-book'
  }))),
  http.post(`${BASE}/v1/shelves`, () => HttpResponse.json(ShelfToJSON({
    id: 'shelf-new',
    path: 'shelf-new',
    description: '',
    domain: '',
    icon: 'i-lucide-book'
  }))),
  http.put(`${BASE}/v1/shelves/:shelfId`, ({ params }) => HttpResponse.json(ShelfToJSON({
    id: params.shelfId as string,
    path: `shelf-${params.shelfId}`,
    description: '',
    domain: '',
    icon: 'i-lucide-book'
  }))),
  http.delete(`${BASE}/v1/shelves/:shelfId`, () => new HttpResponse(null, { status: 204 })),

  // --- Sections ---
  http.get(`${BASE}/v1/sections`, () => HttpResponse.json([])),
  http.post(`${BASE}/v1/sections`, () => HttpResponse.json(SectionToJSON({ id: 'section-new', shelfId: 'shelf-1', title: 'New Section', order: 0 }))),
  http.put(`${BASE}/v1/sections/:sectionId`, ({ params }) => HttpResponse.json(SectionToJSON({ id: params.sectionId as string, shelfId: 'shelf-1', title: 'Updated Section', order: 0 }))),
  http.delete(`${BASE}/v1/sections/:sectionId`, () => new HttpResponse(null, { status: 204 })),
  http.put(`${BASE}/v1/sections/reorder`, () => HttpResponse.json(SectionOrderResponseBodyToJSON({ failures: [] }))),

  // --- Links ---
  http.get(`${BASE}/v1/links`, () => HttpResponse.json([])),
  http.post(`${BASE}/v1/links`, () => HttpResponse.json(LinkToJSON({ id: 'link-new', sectionId: 'section-1', title: 'New Link', link: 'https://example.com', order: 0 }))),
  http.put(`${BASE}/v1/links/:linkId`, ({ params }) => HttpResponse.json(LinkToJSON({ id: params.linkId as string, sectionId: 'section-1', title: 'Updated Link', link: 'https://example.com', order: 0 }))),
  http.delete(`${BASE}/v1/links/:linkId`, () => new HttpResponse(null, { status: 204 })),
  http.put(`${BASE}/v1/links/reorder`, () => HttpResponse.json(LinkOrderResponseBodyToJSON({ failures: [] }))),

  // --- Settings ---
  http.get(`${BASE}/v1/settings`, () => HttpResponse.json(SettingPageBodyToJSON(buildSettingPageBody()))),
  http.put(`${BASE}/v1/settings`, () => HttpResponse.json(SettingPageBodyToJSON(buildSettingPageBody()))),
  http.put(`${BASE}/v1/settings/batch`, () => HttpResponse.json(SettingBatchResponseBodyToJSON({ settings: buildSettingPageBody(), failures: [] }))),
  http.get(`${BASE}/v1/settings/email-delivery`, () => HttpResponse.json({ configured: true, provider: 'smtp' })),

  // --- Themes ---
  http.get(`${BASE}/v1/themes/admin`, () => HttpResponse.json([])),
  http.get(`${BASE}/v1/themes`, () => HttpResponse.json(ThemeGroupedResponseBodyToJSON(buildThemeGrouped()))),
  http.get(`${BASE}/v1/themes/:themeId`, ({ params }) => HttpResponse.json(ThemeToJSON({ id: params.themeId as string, name: 'Test Theme', config: '{}', scope: 'user' }))),
  http.post(`${BASE}/v1/themes`, () => HttpResponse.json(ThemeToJSON({ id: 'theme-new', name: 'New Theme', config: '{}', scope: 'user' }))),
  http.put(`${BASE}/v1/themes/:themeId`, ({ params }) => HttpResponse.json(ThemeToJSON({ id: params.themeId as string, name: 'Updated Theme', config: '{}', scope: 'user' }))),
  http.delete(`${BASE}/v1/themes/:themeId`, () => new HttpResponse(null, { status: 204 })),

  // --- Statistics ---
  http.get(`${BASE}/v1/statistics`, () => HttpResponse.json(StatisticToJSON(buildStatistic())))
]
