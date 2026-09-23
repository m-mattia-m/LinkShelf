const ESCAPE_MAP: Record<string, string> = {
  '&': '&amp;',
  '<': '&lt;',
  '>': '&gt;',
  '"': '&quot;',
  '\'': '&#39;'
}

function escapeHtml(text: string): string {
  return text.replace(/[&<>"']/g, char => ESCAPE_MAP[char]!)
}

// Only http(s) and mailto links are ever turned into a real <a> - anything
// else (javascript:, data:, a bare relative path) stays as plain escaped
// text instead of becoming a link.
const LINK_PATTERN = /\[([^\]]+)]\((https?:\/\/[^\s")]+|mailto:[^\s")]+)\)/g
const BOLD_PATTERN = /\*\*(.+?)\*\*/g
const ITALIC_PATTERN = /\*(.+?)\*/g

/**
 * Renders a small, deliberately restricted subset of Markdown - bold,
 * italic and links only, no headings/lists/images/raw HTML - into safe
 * HTML for v-html. Used for shelf owner-authored text (e.g. the public
 * page footer) shown to anonymous visitors, where full Markdown (which
 * @nuxtjs/mdc's <MDC> renders for admin-authored site content elsewhere in
 * this app) would be more surface area than the feature needs.
 *
 * The input is HTML-escaped first, so the three constructs above are the
 * only tags this can ever produce - anything else the caller typed
 * (including literal HTML) ends up as inert, visible text rather than
 * being interpreted.
 */
export function renderRestrictedMarkdown(text: string): string {
  return escapeHtml(text)
    .replace(LINK_PATTERN, '<a href="$2" target="_blank" rel="noopener noreferrer">$1</a>')
    .replace(BOLD_PATTERN, '<strong>$1</strong>')
    .replace(ITALIC_PATTERN, '<em>$1</em>')
    .replace(/\n/g, '<br>')
}
