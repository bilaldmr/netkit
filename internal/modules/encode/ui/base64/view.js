import { ToolElement, kv } from '/ui/kit.js';

customElements.define('tool-encode-base64', class extends ToolElement {
  static api = '/api/encode-decode/base64';
  static post = true;

  render(v) { return kv([[v.binary ? 'Output (binary, hex)' : 'Output', v.output]]); }
});
