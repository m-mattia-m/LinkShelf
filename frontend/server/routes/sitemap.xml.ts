import { queryCollection } from '@nuxt/content/server'
import { requestHost } from '../../shared/utils/shelfHost'
import { isSiteHost, siteOrigin } from '../../shared/utils/siteSeo'
import { buildSitemap } from '../utils/sitemap'

export default defineEventHandler(async (event) => {
  const origin = siteOrigin(useRuntimeConfig(event).public.siteUrl)

  if (!isSiteHost(origin, requestHost(getRequestHeaders(event)))) {
    throw createError({ statusCode: 404 })
  }

  const docs = await queryCollection(event, 'docs').select('path').all()

  setResponseHeader(event, 'Content-Type', 'application/xml; charset=utf-8')
  setResponseHeader(event, 'Cache-Control', 'public, max-age=3600')
  return buildSitemap(origin, docs.map(doc => doc.path))
})
