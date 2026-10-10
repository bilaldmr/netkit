import { ToolElement, el, kv } from '/static/js/kit.js';

function section(title, rec, issues, extra) {
  const ok = rec && !issues?.length;
  const h = el('h2', null, title + ' ');
  h.append(
    el(
      'span',
      'badge ' + (ok ? 'ok' : rec ? 'warn' : 'bad'),
      ok ? 'OK' : rec ? 'Issues' : 'Missing',
    ),
  );
  const d = el('div');
  d.append(h, kv([['Record', rec], ...(extra || [])]));
  for (const i of issues || []) {
    d.append(el('p', rec ? 'warn' : 'bad', i));
  }
  return d;
}

customElements.define(
  'tool-dns-email',
  class extends ToolElement {
    static api = '/api/dns/email';

    render(v) {
      const d = el('div');
      const tags = Object.entries(v.dmarc.tags || {}).map(([k, x]) => k + '=' + x);
      d.append(
        section('SPF', v.spf.record, v.spf.issues),
        section('DMARC', v.dmarc.record, v.dmarc.issues, [['Tags', tags]]),
        section(
          'DKIM',
          v.dkim.map((k) => k.selector + ': ' + k.record).join('\n'),
          v.dkim.length ? [] : ['No DKIM key found for selectors: ' + v.dkim_checked.join(', ')],
        ),
      );
      return d;
    }
  },
);
