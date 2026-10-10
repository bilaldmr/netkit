import { ToolElement } from '/static/kit.js';

customElements.define('tool-dns-lookup', class extends ToolElement {
  static api = '/api/dns';
});
