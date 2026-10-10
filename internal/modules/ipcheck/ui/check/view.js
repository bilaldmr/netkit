import { ToolElement, el, tree } from '/static/js/kit.js';

customElements.define(
  'tool-ipcheck-check',
  class extends ToolElement {
    static api = '/api/ip';
    static auto = true;

    render(v) {
      const d = el('div'),
        { location: loc, ...rest } = v;
      if (loc) {
        d.append(
          el(
            'p',
            'big',
            [loc.city, loc.region, loc.country].filter(Boolean).join(', ') || 'Unknown location',
          ),
        );
        const p = el('p', 'muted', [loc.org, loc.timezone].filter(Boolean).join(' · ') + ' · ');
        const a = el('a', null, `${loc.lat}, ${loc.lon}`);
        const lat = encodeURIComponent(loc.lat),
          lon = encodeURIComponent(loc.lon);
        a.href = `https://www.openstreetmap.org/?mlat=${lat}&mlon=${lon}#map=11/${lat}/${lon}`;
        a.target = '_blank';
        a.rel = 'noopener';
        p.append(a);
        d.append(p);
      } else if (v.location_error) {
        d.append(el('p', 'error', 'Location unavailable: ' + v.location_error));
      } else if (!v.public) {
        d.append(el('p', 'muted', 'Not a public IP, so no location.'));
      }
      d.append(tree(null, rest));
      return d;
    }
  },
);
