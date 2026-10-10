import { ToolElement, kv } from '/static/kit.js';

customElements.define('tool-encode-hash', class extends ToolElement {
  static api = '/api/encode-decode/hash';
  static post = true;

  render(v) { return kv([['MD5', v.md5], ['SHA-1', v.sha1], ['SHA-256', v.sha256], ['SHA-512', v.sha512]]); }
});
