// Drag and drop functionality for todo reordering
document.addEventListener('DOMContentLoaded', function() {
    const todoList = document.getElementById('todo-list');
    let draggedElement = null;

    // Add drag event listeners to existing todos
    function addDragListeners() {
        const todoItems = document.querySelectorAll('.todo-item');
        todoItems.forEach(item => {
            item.addEventListener('dragstart', handleDragStart);
            item.addEventListener('dragover', handleDragOver);
            item.addEventListener('drop', handleDrop);
            item.addEventListener('dragend', handleDragEnd);
        });
    }

    function handleDragStart(e) {
        draggedElement = this;
        this.classList.add('dragging');
        e.dataTransfer.effectAllowed = 'move';
        e.dataTransfer.setData('text/html', this.outerHTML);
    }

    function handleDragOver(e) {
        e.preventDefault();
        e.dataTransfer.dropEffect = 'move';
        
        const afterElement = getDragAfterElement(todoList, e.clientY);
        if (afterElement == null) {
            todoList.appendChild(draggedElement);
        } else {
            todoList.insertBefore(draggedElement, afterElement);
        }
    }

    function handleDrop(e) {
        e.preventDefault();
        if (draggedElement !== this) {
            // Trigger HTMX reorder
            updateTodoOrder();
        }
    }

    function handleDragEnd(e) {
        this.classList.remove('dragging');
        draggedElement = null;
    }

    function getDragAfterElement(container, y) {
        const draggableElements = [...container.querySelectorAll('.todo-item:not(.dragging)')];
        
        return draggableElements.reduce((closest, child) => {
            const box = child.getBoundingClientRect();
            const offset = y - box.top - box.height / 2;
            
            if (offset < 0 && offset > closest.offset) {
                return { offset: offset, element: child };
            } else {
                return closest;
            }
        }, { offset: Number.NEGATIVE_INFINITY }).element;
    }

    function updateTodoOrder() {
        const todoItems = document.querySelectorAll('.todo-item');
        const todoIds = Array.from(todoItems).map(item => item.dataset.todoId);
        
        // Update hidden inputs
        todoItems.forEach((item, index) => {
            const hiddenInput = item.querySelector('input[name="todo-order"]');
            if (hiddenInput) {
                hiddenInput.value = item.dataset.todoId;
            }
        });

        // Trigger HTMX request
        fetch('/todos/reorder', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded',
            },
            body: 'todo-ids=' + todoIds.join(',')
        });
    }

    // Initialize drag listeners
    addDragListeners();

    // Scroll to last item function
    function scrollToLastItem() {
        const todoItems = document.querySelectorAll('.todo-item');
        if (todoItems.length > 0) {
            const lastItem = todoItems[todoItems.length - 1];
            lastItem.scrollIntoView({ 
                behavior: 'smooth', 
                block: 'center'
            });
            
            // Add a brief highlight effect
            lastItem.classList.add('newly-added');
            setTimeout(() => {
                lastItem.classList.remove('newly-added');
            }, 2000);
        }
    }

    // Handle scroll-to-last trigger
    todoList.addEventListener('scroll-to-last', scrollToLastItem);

    // Edit mode event listeners
    function addEditListeners() {
        // Click on todo text to edit
        document.querySelectorAll('.todo-text.editable').forEach(text => {
            text.addEventListener('click', function() {
                const todoId = this.dataset.todoId;
                editTodo(todoId);
            });
        });

        // Click on edit button
        document.querySelectorAll('.edit-btn').forEach(btn => {
            btn.addEventListener('click', function() {
                const todoId = this.dataset.todoId;
                editTodo(todoId);
            });
        });

        // Click on cancel button
        document.querySelectorAll('.cancel-btn').forEach(btn => {
            btn.addEventListener('click', function() {
                const todoId = this.dataset.todoId;
                cancelEdit(todoId);
            });
        });
    }

    // Initialize edit listeners
    addEditListeners();

    // Re-add listeners when new todos are added via HTMX
    document.body.addEventListener('htmx:afterSwap', function(evt) {
        addDragListeners();
        addEditListeners();
    });
});

// Edit mode functions
function editTodo(todoId) {
    const todoItem = document.getElementById(`todo-${todoId}`);
    const viewMode = todoItem.querySelector('.view-mode');
    const editMode = todoItem.querySelector('.edit-mode');
    const actions = todoItem.querySelector('.todo-actions');
    
    // Hide view mode and actions, show edit mode
    viewMode.style.display = 'none';
    actions.style.display = 'none';
    editMode.style.display = 'block';
    
    // Focus on the text input
    const textInput = editMode.querySelector('input[name="text"]');
    textInput.focus();
    textInput.select();
    
    // Disable dragging while editing
    todoItem.draggable = false;
}

function cancelEdit(todoId) {
    const todoItem = document.getElementById(`todo-${todoId}`);
    const viewMode = todoItem.querySelector('.view-mode');
    const editMode = todoItem.querySelector('.edit-mode');
    const actions = todoItem.querySelector('.todo-actions');
    
    // Show view mode and actions, hide edit mode
    viewMode.style.display = 'block';
    actions.style.display = 'flex';
    editMode.style.display = 'none';
    
    // Re-enable dragging
    todoItem.draggable = true;
    
    // Reset form to original values (form will reset automatically)
    const form = editMode;
    form.reset();
}