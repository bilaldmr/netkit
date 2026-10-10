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
const grades = ['F', 'F', 'F', 'D', 'C', 'B', 'A'];
const inGroup = (g, h) => groups[g].some((m) => (typeof m === 'string' ? m === h : m(h)));

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

function cookieTable(setCookie) {
  const t = el('table', 'data'),
    head = el('tr');
  for (const h of ['Cookie', 'Secure', 'HttpOnly', 'SameSite']) {
    head.append(el('th', null, h));
  }
  t.append(head);
  for (const line of setCookie.split('\n')) {
    const [pair, ...attrs] = line.split(';').map((s) => s.trim());
    const flags = Object.fromEntries(
      attrs.map((a) => {
        const [k, val = ''] = a.split('=');
        return [k.toLowerCase(), val];
      }),
    );
    const flag = (on, text) => el('td', on ? 'ok' : 'bad', text ?? (on ? 'Yes' : 'No'));
    const tr = el('tr');
    tr.append(
      el('td', null, pair.split('=')[0]),
      flag('secure' in flags),
      flag('httponly' in flags),
      flag(!!flags.samesite, flags.samesite || 'Not set'),
    );
    t.append(tr);
  }
  return t;
}

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
        if (g === 'Security' && v.headers['Set-Cookie']) {
          body.append(el('h2', null, 'Cookies'), cookieTable(v.headers['Set-Cookie']));
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
      const present = recommended.filter((h) => h in v.headers).length;
      const grade = grades[present];
      const gradeEl = el('span');
      gradeEl.append(
        el('span', 'badge ' + (grade === 'A' ? 'ok' : grade === 'F' ? 'bad' : 'warn'), grade),
        ` ${present} of ${recommended.length} recommended security headers`,
      );
      d.append(
        kv([
          ['Status', st],
          ['Final URL', v.final],
          ['Protocol', v.proto],
          ['Security grade', gradeEl],
        ]),
        bar,
        body,
      );
      show('All');
      return d;
    }
  },
);
