import { ToolElement, el } from '/static/js/kit.js';

customElements.define(
  'tool-dns-propagation',
  class extends ToolElement {
    static api = '/api/dns/propagation';

    render(v) {
      const d = el('div');
      const agree = v.answers.filter((a) => a.agrees).length;
      const s = el('p');
      s.append(
        el(
          'span',
          'badge ' + (v.consistent ? 'ok' : 'warn'),
          v.consistent ? 'Consistent' : 'Mixed',
        ),
        ' ',
        el(
          'span',
          'muted',
          `${agree}/${v.answers.length} servers return the majority ${v.type} answer`,
        ),
      );
      d.append(s);
      if (!v.consistent) {
        d.append(
          el(
            'p',
            'muted',
            'Different answers right after a change mean propagation is in progress. ' +
              'Large sites and CDNs also return different IPs per resolver on purpose (geo DNS), so mixed results are normal there.',
          ),
        );
      }

      const table = el('table', 'data'),
        head = el('tr');
      for (const h of ['Server', 'Country', 'Answer', 'Status']) head.append(el('th', null, h));
      table.append(head);
      for (const a of v.answers) {
        const tr = el('tr'),
          name = el('td', null, a.server.name + ' ');
        name.append(el('span', 'muted', a.server.addr));
        const recs = el('td', 'records');
        if (a.error) recs.append(el('span', 'error', a.error));
        else for (const r of a.records) recs.append(el('code', null, r));
        const ok = a.agrees
          ? ['ok', '✓ match']
          : a.error
            ? ['bad', '✗ error']
            : ['warn', '≠ differs'];
        const st = el('td', 'status');
        st.append(el('span', 'badge ' + ok[0], ok[1]));
        tr.append(name, el('td', null, a.server.country), recs, st);
        table.append(tr);
      }
      d.append(table);
      return d;
    }
  },
);
