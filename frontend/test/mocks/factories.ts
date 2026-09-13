import type {
  Link,
  Section,
  Setting,
  SettingPageBody,
  Shelf,
  Statistic,
  Theme,
  ThemeGroupedResponseBody,
  TokenPair,
  User
} from '~~/api'

let counter = 0
export function nextId(prefix: string): string {
  counter += 1
  return `${prefix}-${counter}`
}

export function resetFactoryCounter(): void {
  counter = 0
}

export function buildTokenPair(overrides: Partial<TokenPair> = {}): TokenPair {
  return {
    accessToken: 'header.eyJzdWIiOiJ1c2VyLTEiLCJyb2xlIjoidXNlciIsImV4cCI6OTk5OTk5OTk5OX0.signature',
    refreshToken: nextId('refresh-token'),
    ...overrides
  }
}

export function buildAdminTokenPair(overrides: Partial<TokenPair> = {}): TokenPair {
  return buildTokenPair({
    accessToken: 'header.eyJzdWIiOiJ1c2VyLTEiLCJyb2xlIjoiYWRtaW4iLCJleHAiOjk5OTk5OTk5OTl9.signature',
    ...overrides
  })
}

export function buildUser(overrides: Partial<User> = {}): User {
  const id = overrides.id ?? nextId('user')
  return {
    id,
    email: `${id}@example.com`,
    emailVerified: true,
    firstName: 'Jane',
    lastName: 'Doe',
    hasPassword: true,
    role: 'user',
    ...overrides
  } as User
}

export function buildShelf(overrides: Partial<Shelf> = {}): Shelf {
  const id = overrides.id ?? nextId('shelf')
  return {
    id,
    path: `shelf-${id}`,
    description: 'A test shelf',
    domain: '',
    icon: 'i-lucide-book',
    themeId: undefined,
    ...overrides
  } as Shelf
}

export function buildSection(overrides: Partial<Section> = {}): Section {
  return {
    id: nextId('section'),
    shelfId: 'shelf-1',
    title: 'Test Section',
    order: 0,
    ...overrides
  }
}

export function buildLink(overrides: Partial<Link> = {}): Link {
  return {
    id: nextId('link'),
    sectionId: 'section-1',
    title: 'Test Link',
    link: 'https://example.com',
    order: 0,
    ...overrides
  } as Link
}

export function buildTheme(overrides: Partial<Theme> = {}): Theme {
  return {
    id: nextId('theme'),
    name: 'Test Theme',
    config: '{}',
    scope: 'user',
    ownerUserId: 'user-1',
    ...overrides
  } as Theme
}

export function buildThemeGrouped(overrides: Partial<ThemeGroupedResponseBody> = {}): ThemeGroupedResponseBody {
  return {
    instance: [],
    mine: [],
    ...overrides
  }
}

export function buildStatistic(overrides: Partial<Statistic> = {}): Statistic {
  return {
    linkNumber: 3,
    sectionNumber: 2,
    shelfNumber: 1,
    ...overrides
  }
}

export function buildSettingPageBody(overrides: Partial<SettingPageBody> = {}): SettingPageBody {
  return {
    about: '',
    aboutShow: false,
    contact: '',
    contactShow: false,
    imprint: '',
    imprintShow: false,
    termsOfUse: '',
    termsOfUseShow: false,
    privacyPolicy: '',
    privacyPolicyShow: false,
    redirectToDashboard: false,
    ...overrides
  } as SettingPageBody
}

export function buildSetting(overrides: Partial<Setting> = {}): Setting {
  return {
    key: 'about',
    languageCode: 'en',
    value: '',
    ...overrides
  }
}
