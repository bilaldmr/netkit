import { ToolElement, kv } from '/static/kit.js';

customElements.define('tool-encode-url', class extends ToolElement {
  static api = '/api/encode-decode/url';
  static post = true;

  render(v) {
    return kv([['Query component', v.query], ['Path segment', v.path], ['Decoded', v.text]]);
  }
});
