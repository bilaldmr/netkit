import { el } from '/static/js/kit.js';

export function copyList(items) {
  const ul = el('ul', 'list');
  for (const s of items) {
    const li = el('li'),
      b = el('button', 'copy secondary', 'Copy');
    b.type = 'button';
    b.onclick = async () => {
      await navigator.clipboard.writeText(s);
      b.textContent = 'Copied';
      setTimeout(() => (b.textContent = 'Copy'), 1200);
    };
    li.append(el('code', null, s), b);
    ul.append(li);
  }
  return ul;
}
