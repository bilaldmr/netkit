import { ToolElement, el } from '/static/js/kit.js';

customElements.define(
  'tool-sitefiles-files',
  class extends ToolElement {
    static api = '/api/site-files';

    render(v) {
      const d = el('div');
      for (const f of v.files) {
        const h = el('h2', null, f.path + ' ');
        h.append(
          f.error
            ? el('span', 'badge bad', 'Error')
            : el('span', 'badge s' + String(f.status)[0], f.status),
        );
        const link = el('a', null, f.url);
        link.href = f.url;
        link.target = '_blank';
        link.rel = 'noopener noreferrer';
        const meta = el('p', 'muted');
        meta.append(link);
        if (f.path === '/sitemap.xml' && f.body) {
          meta.append(` · ${(f.body.match(/<loc>/g) || []).length} URLs`);
        }
        d.append(h, meta);
        // Many sites answer missing files with their HTML error page and a 200 status.
        const html = f.content_type?.startsWith('text/html');
        if (f.error) d.append(el('p', 'error', f.error));
        else if (html) d.append(el('p', 'warn', 'Not found: the site returned an HTML page.'));
        else if (f.body) d.append(el('pre', null, f.body));
        else if (f.status !== 200) d.append(el('p', 'muted', 'Not found.'));
        if (f.truncated && !html) d.append(el('p', 'warn', 'Showing the first 256 KB.'));
      }
      return d;
    }
  },
);
