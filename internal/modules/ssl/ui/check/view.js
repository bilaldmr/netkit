import { ToolElement, el, tree } from '/static/js/kit.js';

customElements.define('tool-ssl-check', class extends ToolElement {
  static api = '/api/ssl';

  render(v) {
    const d = el('div'), s = el('p');
    const days = v.days_left < 0 ? 'bad' : v.days_left < 14 ? 'warn' : 'ok';
    s.append(el('span', 'badge ' + (v.valid ? 'ok' : 'bad'), v.valid ? 'Valid' : 'Invalid'), ' ',
             el('span', 'badge ' + days, v.days_left + ' days left'), ' ',
             el('span', 'muted', `${v.tls_version} · ${v.cipher}`));
    d.append(s);
    if (v.verify_error) d.append(el('p', 'error', v.verify_error));
    d.append(el('h2', null, 'Certificate chain'), tree(null, v.chain));
    return d;
  }
});
