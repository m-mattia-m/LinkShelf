import { requestHost } from '../../shared/utils/shelfHost'
import { isSiteHost, siteOrigin } from '../../shared/utils/siteSeo'
import { buildRobots } from '../utils/robots'

export default defineEventHandler((event) => {
  const origin = siteOrigin(useRuntimeConfig(event).public.siteUrl)

  setResponseHeader(event, 'Content-Type', 'text/plain; charset=utf-8')
  return buildRobots(origin, isSiteHost(origin, requestHost(getRequestHeaders(event))))
})
