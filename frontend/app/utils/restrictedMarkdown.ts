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

// Only http(s) and mailto links become an <a>; anything else stays plain text.
const LINK_PATTERN = /\[([^\]]+)]\((https?:\/\/[^\s")]+|mailto:[^\s")]+)\)/g
const BOLD_PATTERN = /\*\*(.+?)\*\*/g
const ITALIC_PATTERN = /\*(.+?)\*/g

/** Renders bold, italic and links as HTML. The input is escaped first. */
export function renderRestrictedMarkdown(text: string): string {
  return escapeHtml(text)
    .replace(LINK_PATTERN, '<a href="$2" target="_blank" rel="noopener noreferrer">$1</a>')
    .replace(BOLD_PATTERN, '<strong>$1</strong>')
    .replace(ITALIC_PATTERN, '<em>$1</em>')
    .replace(/\n/g, '<br>')
}
