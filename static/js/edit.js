// Initialize sidebar data
window.sidebarData = {
    currentPath: window.currentPath || "",
    folderPath: window.folderPath || "",
    noteTitle: window.noteTitle || ""
};

// Initialize editor
let editor;
let isDirty = false;
let autoSaveInterval;
let lastSavedContent = '';

document.addEventListener('DOMContentLoaded', function() {
    // Initialize CodeMirror
    editor = CodeMirror.fromTextArea(document.getElementById('content'), {
        mode: 'markdown',
        theme: 'monokai',
        lineNumbers: true,
        lineWrapping: true,
        autofocus: true,
        extraKeys: {
            "Tab": "indentMore",
            "Shift-Tab": "indentLess",
            "Ctrl-S": function(cm) {
                saveContent();
            }
        }
    });

    // Set initial content
    const initialContent = document.getElementById('content').value;
    editor.setValue(initialContent);
    lastSavedContent = initialContent;

    // Handle content changes
    editor.on('change', function() {
        isDirty = true;
        updateSaveButton();
    });

    // Setup auto-save
    setupAutoSave();

    // Setup form submission
    setupFormSubmission();

    // Setup keyboard shortcuts
    setupKeyboardShortcuts();

    // Setup preview functionality
    setupPreview();

    // Setup sidebar
    setupSidebar();
});

function setupAutoSave() {
    // Auto-save every 30 seconds if content is dirty
    autoSaveInterval = setInterval(function() {
        if (isDirty) {
            saveContent();
        }
    }, 30000);
}

function setupFormSubmission() {
    const form = document.getElementById('note-form');
    form.addEventListener('submit', function(e) {
        e.preventDefault();
        saveContent();
    });
}

function setupKeyboardShortcuts() {
    // Save shortcut (Ctrl/Cmd + S)
    document.addEventListener('keydown', function(e) {
        if ((e.ctrlKey || e.metaKey) && e.key === 's') {
            e.preventDefault();
            saveContent();
        }
    });

    // Preview shortcut (Ctrl/Cmd + P)
    document.addEventListener('keydown', function(e) {
        if ((e.ctrlKey || e.metaKey) && e.key === 'p') {
            e.preventDefault();
            togglePreview();
        }
    });
}

function setupPreview() {
    const previewBtn = document.getElementById('preview-btn');
    const previewContent = document.getElementById('preview-content');
    const editorContainer = document.querySelector('.editor-container');
    const previewContainer = document.getElementById('preview-container');

    previewBtn.addEventListener('click', togglePreview);

    function togglePreview() {
        const isPreviewVisible = previewContainer.style.display === 'block';
        
        if (isPreviewVisible) {
            // Switch back to editor
            previewContainer.style.display = 'none';
            editorContainer.style.display = 'block';
            previewBtn.innerHTML = '<i class="fas fa-eye"></i> Preview';
        } else {
            // Switch to preview
            const content = editor.getValue();
            previewContent.innerHTML = marked.parse(content);
            editorContainer.style.display = 'none';
            previewContainer.style.display = 'block';
            previewBtn.innerHTML = '<i class="fas fa-edit"></i> Edit';
        }
    }
}

function setupSidebar() {
    // Initialize tree expand/collapse
    setupTreeExpandCollapse();
    
    // Auto-expand path to current folder
    expandPathToCurrentFolder(window.currentPath);
}

function saveContent() {
    const content = editor.getValue();
    const title = document.getElementById('title').value;
    const folder = document.getElementById('folder').value;
    
    // Show saving indicator
    const saveBtn = document.getElementById('save-btn');
    const originalText = saveBtn.innerHTML;
    saveBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Saving...';
    saveBtn.disabled = true;
    
    // Create the request body
    const requestBody = JSON.stringify({ 
        title: title,
        content: content,
        folder: folder
    });
    
    // Send request to save note
    fetch('/save', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: requestBody
    })
    .then(response => {
        if (!response.ok) {
            throw new Error('Server returned ' + response.status);
        }
        return response.json();
    })
    .then(data => {
        console.log('Success:', data);
        isDirty = false;
        lastSavedContent = content;
        updateSaveButton();
        
        // Update URL if title changed
        if (data.redirect) {
            window.history.pushState({}, '', data.redirect);
        }
        
        // Show success message
        showNotification('Note saved successfully', 'success');
    })
    .catch(error => {
        console.error('Error:', error);
        showNotification('Error saving note: ' + error.message, 'error');
    })
    .finally(() => {
        // Restore save button
        saveBtn.innerHTML = originalText;
        saveBtn.disabled = false;
    });
}

function updateSaveButton() {
    const saveBtn = document.getElementById('save-btn');
    if (isDirty) {
        saveBtn.classList.add('dirty');
        saveBtn.innerHTML = '<i class="fas fa-save"></i> Save Changes';
    } else {
        saveBtn.classList.remove('dirty');
        saveBtn.innerHTML = '<i class="fas fa-check"></i> Saved';
    }
}

function showNotification(message, type = 'info') {
    const notification = document.createElement('div');
    notification.className = `notification ${type}`;
    notification.innerHTML = `
        <i class="fas ${type === 'success' ? 'fa-check-circle' : 'fa-exclamation-circle'}"></i>
        ${message}
    `;
    
    document.body.appendChild(notification);
    
    // Remove notification after 3 seconds
    setTimeout(() => {
        notification.remove();
    }, 3000);
}

// Handle page unload
window.addEventListener('beforeunload', function(e) {
    if (isDirty) {
        e.preventDefault();
        e.returnValue = '';
    }
}); 