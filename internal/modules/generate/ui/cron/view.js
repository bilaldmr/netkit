import { ToolElement, el, kv } from '/static/js/kit.js';

customElements.define(
  'tool-generate-cron',
  class extends ToolElement {
    static api = '/api/generate/cron';

    render(v) {
      const d = el('div');
      d.append(
        kv(v.fields.map((f) => [f.name, f.values])),
        el('h2', null, `Next runs (${v.timezone})`),
        kv(v.next.map((t, i) => [String(i + 1), t])),
      );
      return d;
    }
  },
);
