export const el = (tag, cls, text) => {
  const e = document.createElement(tag);
  if (cls) e.className = cls;
  if (text != null) e.textContent = text;
  return e;
};

export function tree(key, v) {
  const label = key == null ? [] : [el('span', 'k', JSON.stringify(key)), el('span', 'muted', ': ')];
  if (v !== null && typeof v === 'object') {
    const arr = Array.isArray(v), entries = Object.entries(v);
    const d = el('details'); d.open = true;
    const s = el('summary');
    s.append(...label, el('span', 'muted', arr ? `[${entries.length}]` : `{${entries.length}}`));
    d.append(s, ...entries.map(([k, x]) => tree(arr ? null : k, x)));
    if (key == null) { const root = el('div', 'tree'); root.append(d); return root; }
    return d;
  }
  const leaf = el('div', 'leaf' + (key == null ? ' tree' : ''));
  leaf.append(...label, el('span', 't-' + ({ string: 's', number: 'n', boolean: 'b' }[typeof v] || 'z'), JSON.stringify(v)));
  return leaf;
}

export async function api(url, init) {
  const r = await fetch(url, init);
  const v = await r.json().catch(() => ({ error: `HTTP ${r.status}` }));
  if (!r.ok) throw new Error(v.error || `HTTP ${r.status}`);
  return v;
}

export function kv(rows) {
  const t = el('table', 'kv');
  for (const [k, v] of rows) {
    if (v == null || v === '' || (Array.isArray(v) && !v.length)) continue;
    const tr = el('tr'), td = el('td');
    if (v instanceof Node) td.append(v);
    else td.textContent = Array.isArray(v) ? v.join('\n') : String(v);
    tr.append(el('th', null, k), td);
    t.append(tr);
  }
  return t;
}

export class ToolElement extends HTMLElement {
  static api = '';
  static auto = false;
  static post = false;

  connectedCallback() {
    if (this.out) return;
    const root = this.shadowRoot;
    this.formEl = root.querySelector('form');
    this.out = root.querySelector('.out');
    this.err = root.querySelector('.err');
    this.setup();
  }

  setup() {
    const { api: path, auto } = this.constructor, form = this.formEl;
    if (this.constructor.post) {
      if (form) form.onsubmit = e => {
        e.preventDefault();
        this.load(path, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(Object.fromEntries(new FormData(form))),
        });
      };
      return;
    }
    const q = new URLSearchParams(location.search);
    const fields = () => {
      const p = new URLSearchParams();
      if (form) for (const [k, v] of new FormData(form)) if (v) p.set(k, v);
      return p;
    };
    if (form) {
      for (const [k, v] of q) { const f = form.elements.namedItem(k); if (f) f.value = v; }
      form.onsubmit = e => {
        e.preventDefault();
        const p = fields();
        history.replaceState(null, '', p.size ? '?' + p : location.pathname);
        this.load(`${path}?${p}`);
      };
    }
    if (auto || q.size) this.load(`${path}?${fields()}`);
  }

  async load(url, init) {
    this.err.textContent = '';
    this.out.replaceChildren(el('p', 'muted', 'Loading…'));
    try { this.out.replaceChildren(this.render(await api(url, init))); }
    catch (e) { this.out.replaceChildren(); this.err.textContent = e.message; }
  }

  render(v) { return tree(null, v); }
}
