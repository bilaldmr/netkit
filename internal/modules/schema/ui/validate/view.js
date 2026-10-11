import { ToolElement, el, kv, tree } from '/static/js/kit.js';

function issueTable(issues) {
  const t = el('table', 'data'),
    head = el('tr');
  for (const h of ['Level', 'Source', 'Path', 'Issue']) head.append(el('th', null, h));
  t.append(head);
  for (const i of issues) {
    const tr = el('tr');
    tr.append(
      el('td', i.level === 'error' ? 'bad' : 'warn', i.level),
      el('td', null, i.source === 'google' ? 'Google rich results' : 'schema.org'),
      el('td', null, i.path),
      el('td', null, i.msg),
    );
    t.append(tr);
  }
  return t;
}

customElements.define(
  'tool-schema-validate',
  class extends ToolElement {
    static api = '/api/schema';

    render(v) {
      const d = el('div');
      const errors = v.blocks.flatMap((b) => b.issues).filter((i) => i.level === 'error').length;
      const warnings = v.blocks.flatMap((b) => b.issues).length - errors;
      d.append(
        kv([
          ['Status', el('span', 'badge s' + String(v.status)[0], v.status)],
          ['Final URL', v.final],
          ['JSON-LD blocks', String(v.blocks.length)],
          ['Errors', el('span', errors ? 'bad' : 'ok', String(errors))],
          ['Warnings', el('span', warnings ? 'warn' : 'ok', String(warnings))],
        ]),
      );
      if (!v.blocks.length) d.append(el('p', 'warn', 'No application/ld+json script found.'));
      v.blocks.forEach((b, i) => {
        const h = el('h2', null, `#${i + 1} ${b.types?.join(', ') || 'Untyped'} `);
        const where = `script ${b.script}` + (b.in_graph ? ` · ${b.path.slice(2)}` : '');
        h.append(el('span', 'badge', where));
        d.append(h);
        d.append(
          b.issues.length ? issueTable(b.issues) : el('p', 'ok', 'Valid against schema.org.'),
        );
        if (b.raw) d.append(el('pre', null, b.raw));
        else d.append(tree(null, b.data));
      });
      return d;
    }
  },
);
