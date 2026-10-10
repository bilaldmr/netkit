import { ToolElement, el } from '/static/js/kit.js';

customElements.define('tool-redirects-trace', class extends ToolElement {
  static api = '/api/redirects';

  render(v) {
    const d = el('div'), ol = el('ol');
    for (const h of v.hops || []) {
      const li = el('li');
      li.append(el('span', 'badge s' + String(h.status)[0], h.status), ' ', el('code', null, h.url));
      if (h.server) li.append(' ', el('span', 'muted', h.server));
      ol.append(li);
    }
    d.append(ol);
    if (v.error) d.append(el('p', 'error', v.error));
    else if (v.final) { const p = el('p', null, 'Final: '); p.append(el('code', null, v.final)); d.append(p); }
    return d;
  }
});
