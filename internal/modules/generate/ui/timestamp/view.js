import { ToolElement, kv } from '/static/js/kit.js';

customElements.define(
  'tool-generate-timestamp',
  class extends ToolElement {
    static api = '/api/generate/timestamp';
    static auto = true;

    render(v) {
      return kv([
        ['Unix (s)', v.unix],
        ['Unix (ms)', v.unix_ms],
        ['UTC', v.utc],
        [v.timezone, v.local],
        ['RFC 1123', v.rfc1123],
        ['Relative', v.relative],
      ]);
    }
  },
);
