import { ToolElement, el, api } from '/static/js/kit.js';

customElements.define('tool-myip-info', class extends ToolElement {
  async setup() {
    const root = this.shadowRoot;
    const here = api('/api/myip');
    const probes = await api('/api/myip/probes').catch(() => ({}));

    const detect = async v => {
      if (probes[v]) {
        try {
          return { ip: (await api(probes[v] + '/api/myip', { signal: AbortSignal.timeout(5000) })).ip };
        } catch {
          return { none: 'Not available' };
        }
      }
      const h = await here.catch(() => null);
      return h && 'v' + h.version === v ? { ip: h.ip } : { none: 'Not checked', unchecked: true };
    };

    const results = await Promise.all([detect('v4'), detect('v6')]);
    let unchecked = false;
    for (const [v, r] of [['v4', results[0]], ['v6', results[1]]]) {
      const slot = root.querySelector(`[data-v="${v}"]`);
      slot.textContent = r.ip || r.none;
      slot.classList.toggle('none', !r.ip);
      unchecked ||= !!r.unchecked;
    }

    try {
      root.querySelector('.ua').textContent = (await here).user_agent;
    } catch (e) {
      this.err.textContent = e.message;
    }
    if (unchecked) {
      this.out.replaceChildren(el('p', 'muted',
        'Only the address of this connection is visible. Set MYIP_V4_URL and MYIP_V6_URL to detect both.'));
    }
  }
});
