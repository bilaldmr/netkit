import { ToolElement, el, kv } from '/ui/kit.js';

customElements.define('tool-cidr-calc', class extends ToolElement {
  static api = '/api/cidr';

  render(v) {
    const d = el('div');
    if (v.contains != null) {
      const p = el('p');
      p.append(el('span', 'badge ' + (v.contains ? 'ok' : 'bad'), v.contains ? 'Inside' : 'Outside'), ' ',
               el('span', 'muted', `${v.ip} is ${v.contains ? '' : 'not '}in ${v.cidr}`));
      d.append(p);
    }
    d.append(kv([
      ['CIDR', v.cidr], ['Network', v.network], ['Broadcast', v.broadcast],
      ['Netmask', v.netmask], ['Wildcard', v.wildcard],
      ['First host', v.first_host], ['Last host', v.last_host],
      ['Addresses', v.addresses], ['Usable hosts', v.hosts], ['Version', 'IPv' + v.version],
    ]));
    return d;
  }
});
