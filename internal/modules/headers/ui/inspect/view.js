import { ToolElement, el, kv } from '/static/js/kit.js';

const groups = {
  Security: [
    'strict-transport-security',
    'content-security-policy',
    'content-security-policy-report-only',
    'x-frame-options',
    'x-content-type-options',
    'referrer-policy',
    'cross-origin-opener-policy',
    'cross-origin-embedder-policy',
    'cross-origin-resource-policy',
    'x-xss-protection',
    'x-permitted-cross-domain-policies',
  ],
  Permissions: ['permissions-policy', 'feature-policy', 'document-policy'],
  Cache: [
    'cache-control',
    'expires',
    'etag',
    'last-modified',
    'age',
    'pragma',
    'vary',
    'cdn-cache-control',
    'surrogate-control',
    'x-cache',
    'cf-cache-status',
    'x-cache-status',
  ],
  CORS: [(h) => h.startsWith('access-control-'), 'timing-allow-origin'],
};
const recommended = [
  'Strict-Transport-Security',
  'Content-Security-Policy',
  'X-Frame-Options',
  'X-Content-Type-Options',
  'Referrer-Policy',
  'Permissions-Policy',
];
const inGroup = (g, h) => groups[g].some((m) => (typeof m === 'string' ? m === h : m(h)));

// Long policy headers are lists; one directive per line reads better.
const pretty = (k, v) => {
  k = k.toLowerCase();
  if (k.startsWith('content-security-policy'))
    return v
      .split(';')
      .map((s) => s.trim())
      .filter(Boolean)
      .join('\n');
  if (k === 'permissions-policy' || k === 'feature-policy') return v.split(/,\s*/).join('\n');
  return v;
};

customElements.define(
  'tool-headers-inspect',
  class extends ToolElement {
    static api = '/api/headers';

    render(v) {
      const d = el('div'),
        bar = el('div', 'filters'),
        body = el('div');
      const keys = Object.keys(v.headers).sort();
      const show = (g) => {
        for (const b of bar.children) b.setAttribute('aria-pressed', b.textContent.startsWith(g));
        const list = g === 'All' ? keys : keys.filter((k) => inGroup(g, k.toLowerCase()));
        body.replaceChildren(
          list.length
            ? kv(list.map((k) => [k, pretty(k, v.headers[k])]))
            : el('p', 'muted', 'No headers in this group.'),
        );
        if (g === 'Security' || g === 'Permissions') {
          const missing = recommended.filter(
            (h) => inGroup(g, h.toLowerCase()) && !(h in v.headers),
          );
          if (missing.length) body.append(el('p', 'warn', 'Missing: ' + missing.join(', ')));
        }
      };
      for (const g of ['All', ...Object.keys(groups)]) {
        const n =
          g === 'All' ? keys.length : keys.filter((k) => inGroup(g, k.toLowerCase())).length;
        const b = el('button', null, `${g} (${n})`);
        b.type = 'button';
        b.onclick = () => show(g);
        bar.append(b);
      }
      const st = el('span', 'badge s' + String(v.status)[0], v.status);
      d.append(
        kv([
          ['Status', st],
          ['Final URL', v.final],
          ['Protocol', v.proto],
        ]),
        bar,
        body,
      );
      show('All');
      return d;
    }
  },
);
