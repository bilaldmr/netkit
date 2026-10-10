import { ToolElement, el } from '/ui/kit.js';

const badge = { open: 'ok', closed: 'muted', filtered: 'warn', error: 'bad' };

customElements.define('tool-ports-check', class extends ToolElement {
  static api = '/api/ports';

  render(v) {
    const d = el('div');
    const open = v.results.filter(r => r.state === 'open').length;
    d.append(el('p', 'muted', `${v.host} (${v.address}): ${open} of ${v.results.length} ports open`));
    const table = el('table'), head = el('tr');
    for (const h of ['Port', 'Service', 'State', 'Latency']) head.append(el('th', null, h));
    table.append(head);
    for (const r of v.results) {
      const tr = el('tr'), st = el('td');
      st.append(el('span', 'badge ' + badge[r.state], r.state));
      if (r.error) st.append(' ', el('span', 'muted', r.error));
      tr.append(el('td', null, r.port), el('td', null, r.service || ''), st,
                el('td', null, r.latency_ms ? r.latency_ms.toFixed(1) + ' ms' : ''));
      table.append(tr);
    }
    d.append(table);
    return d;
  }
});
