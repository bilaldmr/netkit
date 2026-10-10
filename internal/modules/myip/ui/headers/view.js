import { ToolElement, el, kv } from '/ui/kit.js';

customElements.define('tool-myip-headers', class extends ToolElement {
  static api = '/api/myip/headers';
  static auto = true;

  render(v) {
    const d = el('div');
    d.append(
      kv([['Client IP', v.client_ip], ['Remote address', v.remote_addr], ['Method', v.method], ['Protocol', v.proto], ['Host', v.host]]),
      el('h2', null, 'Headers'),
      kv(Object.keys(v.headers).sort().map(k => [k, v.headers[k]])),
    );
    return d;
  }
});
