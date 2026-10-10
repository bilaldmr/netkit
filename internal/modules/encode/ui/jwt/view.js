import { ToolElement, el, kv, tree } from '/static/js/kit.js';

customElements.define('tool-encode-jwt', class extends ToolElement {
  static api = '/api/encode-decode/jwt';
  static post = true;

  render(v) {
    const d = el('div');
    if (v.expired != null) {
      const p = el('p');
      p.append(el('span', 'badge ' + (v.expired ? 'bad' : 'ok'), v.expired ? 'Expired' : 'Not expired'));
      d.append(p);
    }
    d.append(
      kv([['Algorithm', v.algorithm], ['Issued at', v.issued_at], ['Not before', v.not_before], ['Expires at', v.expires_at]]),
      el('h2', null, 'Header'), tree(null, v.header),
      el('h2', null, 'Payload'), tree(null, v.payload),
      el('h2', null, 'Signature'), kv([['Base64url', v.signature]]),
    );
    return d;
  }
});
