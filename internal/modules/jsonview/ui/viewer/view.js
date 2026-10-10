import { ToolElement } from '/static/js/kit.js';

customElements.define(
  'tool-jsonview-viewer',
  class extends ToolElement {
    setup() {
      const form = this.formEl;
      const fromURL = (u) => this.load('/api/json?url=' + encodeURIComponent(u));
      const fromData = (data) => this.load('/api/json', { method: 'POST', body: data });
      form.onsubmit = (e) => {
        e.preventDefault();
        const url = form.elements.url.value.trim(),
          data = form.elements.data.value.trim();
        if (data) {
          history.replaceState(null, '', location.pathname);
          fromData(data);
        } else if (url) {
          history.replaceState(null, '', '?url=' + encodeURIComponent(url));
          fromURL(url);
        }
      };
      for (const b of form.querySelectorAll('[data-open]')) {
        b.onclick = () =>
          this.out.querySelectorAll('details').forEach((d) => (d.open = !!b.dataset.open));
      }
      const q = new URLSearchParams(location.search),
        u = q.get('url'),
        data = q.get('data');
      if (u) {
        form.elements.url.value = u;
        fromURL(u);
      } else if (data) {
        form.elements.data.value = data;
        fromData(data);
      }
    }
  },
);
