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

    // Re-add listeners when new todos are added via HTMX
    document.body.addEventListener('htmx:afterSwap', function(evt) {
        addDragListeners();
    });
});