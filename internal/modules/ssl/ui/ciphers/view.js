import { ToolElement, el } from '/static/js/kit.js';

const yes = (on) => el('td', on ? 'ok' : 'muted', on ? 'Yes' : '–');

function table(headers, rows) {
  const t = el('table', 'data'),
    head = el('tr');
  for (const h of headers) {
    head.append(el('th', null, h));
  }
  t.append(head, ...rows);
  return t;
}

customElements.define(
  'tool-ssl-ciphers',
  class extends ToolElement {
    static api = '/api/ssl/ciphers';

    render(v) {
      const d = el('div'),
        hosts = v.hosts,
        names = hosts.map((h) => (h.port === '443' ? h.host : `${h.host}:${h.port}`));

      for (const [i, h] of hosts.entries()) {
        if (h.error) d.append(el('p', 'error', `${names[i]}: ${h.error}`));
      }

      if (hosts.length === 2 && !hosts.some((h) => h.error)) {
        const ok = v.common_ciphers?.length > 0;
        const p = el('p');
        p.append(
          el('span', 'badge ' + (ok ? 'ok' : 'bad'), ok ? 'Compatible' : 'No common cipher'),
          ' ',
          ok
            ? `${v.common_ciphers.length} ciphers over ${v.common_versions.join(', ')}`
            : 'these hosts share no TLS version and cipher pair',
        );
        d.append(p);
      }

      d.append(
        el('h2', null, 'TLS versions'),
        table(
          ['Version', ...names],
          hosts[0].versions.map((ver, i) => {
            const tr = el('tr');
            tr.append(
              el('td', null, ver.version),
              ...hosts.map((h) => yes(h.versions[i].supported)),
            );
            return tr;
          }),
        ),
      );

      const rows = [];
      for (const [i, ver] of hosts[0].versions.entries()) {
        const all = [...new Set(hosts.flatMap((h) => h.versions[i].ciphers || []))].sort();
        for (const c of all) {
          const tr = el('tr'),
            name = el('td');
          name.append(el('code', null, c));
          if (v.weak.includes(c)) name.append(' ', el('span', 'badge warn', 'weak'));
          tr.append(
            el('td', null, ver.version),
            name,
            ...hosts.map((h) => yes((h.versions[i].ciphers || []).includes(c))),
          );
          rows.push(tr);
        }
      }
      d.append(
        el('h2', null, 'Ciphers'),
        rows.length
          ? table(['Version', 'Cipher', ...names], rows)
          : el('p', 'muted', 'None found.'),
        el(
          'p',
          'muted',
          "Scanned with Go's TLS client: TLS 1.3 shows only the cipher the server picked, DHE, CCM, ARIA and Camellia suites are not tested, and a few servers answer it differently than OpenSSL or Java clients.",
        ),
      );
      return d;
    }
  },
);
