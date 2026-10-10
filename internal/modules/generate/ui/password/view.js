import { ToolElement } from '/static/kit.js';
import { copyList } from '../list.js';

customElements.define('tool-generate-password', class extends ToolElement {
  static api = '/api/generate/password';
  static auto = true;

  render(v) { return copyList(v.passwords); }
});
