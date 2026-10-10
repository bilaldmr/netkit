import { ToolElement, el, kv, tree } from '/static/kit.js';

const date = s => (s ? s.slice(0, 10) : '');

customElements.define('tool-whois-lookup', class extends ToolElement {
  static api = '/api/whois';

  render(v) {
    const d = el('div');
    if (v.days_left != null) {
      const cls = v.days_left < 0 ? 'bad' : v.days_left < 30 ? 'warn' : 'ok';
      const p = el('p');
      p.append(el('span', 'badge ' + cls, v.days_left < 0 ? 'Expired' : `Expires in ${v.days_left} days`));
      d.append(p);
    }
    d.append(kv([
      ['Name', v.name], ['Handle', v.handle], ['Range', v.range], ['Country', v.country],
      ['Registrar', v.registrar], ['Registrant', v.registrant],
      ['Registered', date(v.registered)], ['Updated', date(v.updated)], ['Expires', date(v.expires)],
      ['Status', v.status], ['Nameservers', v.nameservers],
      ['DNSSEC', v.dnssec == null ? null : v.dnssec ? 'signed' : 'unsigned'],
      ['RDAP server', v.server],
    ]));
    const raw = el('details'), s = el('summary', 'muted', 'Raw RDAP response');
    raw.append(s, tree(null, v.raw));
    d.append(raw);
    return d;
  }
});
