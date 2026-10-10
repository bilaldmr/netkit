import { ToolElement } from '/static/kit.js';
import { copyList } from '../list.js';

customElements.define('tool-generate-uuid', class extends ToolElement {
  static api = '/api/generate/uuid';
  static auto = true;

  render(v) { return copyList(v.uuids); }
});
