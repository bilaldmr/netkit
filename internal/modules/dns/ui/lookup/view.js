import { ToolElement } from '/static/js/kit.js';

customElements.define(
  'tool-dns-lookup',
  class extends ToolElement {
    static api = '/api/dns';
  },
);
