const todoList = document.getElementById('todo-list');
const todoForm = document.getElementById('todo-form');
const todoTitle = document.getElementById('todo-title');

async function fetchTodos() {
  const res = await fetch('/api/todos');
  const todos = await res.json();

  todoList.innerHTML = '';
  todos.forEach((todo) => {
    const li = document.createElement('li');

    const text = document.createElement('span');
    text.textContent = todo.title;
    if (todo.completed) {
      text.classList.add('done');
    }

    const actions = document.createElement('div');

    const toggleButton = document.createElement('button');
    toggleButton.textContent = todo.completed ? '戻す' : '完了';
    toggleButton.onclick = async () => {
      await fetch(`/api/todos/${todo.id}/toggle`, { method: 'PUT' });
      fetchTodos();
    };

    const deleteButton = document.createElement('button');
    deleteButton.textContent = '削除';
    deleteButton.onclick = async () => {
      await fetch(`/api/todos/${todo.id}`, { method: 'DELETE' });
      fetchTodos();
    };

    actions.appendChild(toggleButton);
    actions.appendChild(deleteButton);

    li.appendChild(text);
    li.appendChild(actions);
    todoList.appendChild(li);
  });
}

todoForm.addEventListener('submit', async (event) => {
  event.preventDefault();
  const title = todoTitle.value.trim();
  if (!title) return;

  await fetch('/api/todos', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ title }),
  });

  todoTitle.value = '';
  fetchTodos();
});

fetchTodos();
