// Initialize sidebar data
window.sidebarData = {
    currentPath: window.currentPath || "",
    folderPath: window.folderPath || "",
    noteTitle: window.noteTitle || ""
};

// Initialize editor
let isDirty = false;
let lastSavedContent = '';

// Initialize everything when the DOM is ready
document.addEventListener('DOMContentLoaded', function() {
    console.log('DOM Content Loaded');
    initializeEditor();
    initializeEventListeners();
});

function initializeEditor() {
    console.log('Initializing editor');
    const editor = document.getElementById('editor');
    
    // Set initial content
    lastSavedContent = editor.value;

    // Handle content changes
    editor.addEventListener('input', function() {
        isDirty = true;
        updateSaveButton();
    });
}

function initializeEventListeners() {
    console.log('Initializing event listeners');
    // Setup form submission
    setupFormSubmission();

    // Setup keyboard shortcuts
    setupKeyboardShortcuts();

    // Setup sidebar
    setupSidebar();
}

function setupFormSubmission() {
    console.log('Setting up form submission');
    const form = document.getElementById('note-form');
    const saveBtn = document.getElementById('save-btn');
    
    if (!form || !saveBtn) {
        console.error('Form or save button not found');
        return;
    }
    
    // Handle form submission
    form.addEventListener('submit', function(e) {
        e.preventDefault();
        console.log('Form submitted');
        saveContent();
    });

    // Handle save button click
    saveBtn.addEventListener('click', function(e) {
        e.preventDefault();
        console.log('Save button clicked');
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
}

function setupSidebar() {
    // Initialize tree expand/collapse
    setupTreeExpandCollapse();
    
    // Auto-expand path to current folder
    expandPathToCurrentFolder(window.currentPath);
}

function saveContent() {
    const editor = tinymce.get('editor');
    const titleInput = document.getElementById('title');
    const folderPathInput = document.querySelector('input[name="folder_path"]');
    const originalTitleInput = document.querySelector('input[name="original_title"]');

    if (!editor || !titleInput || !folderPathInput) {
        console.error('Required form elements not found');
        return;
    }

    // Get content from TinyMCE editor
    const content = editor.getContent();
    const title = titleInput.value;
    const folderPath = folderPathInput.value;
    const oldTitle = originalTitleInput ? originalTitleInput.value : '';

    // Debug logging
    console.log('=== Debug: Content Flow ===');
    console.log('1. TinyMCE editor instance:', editor);
    console.log('2. Editor content:', content);
    console.log('3. Content length:', content.length);
    console.log('4. Content type:', typeof content);
    console.log('5. Is content empty?', !content);
    console.log('6. Title:', title);
    console.log('7. Old title:', oldTitle);
    console.log('8. Folder path:', folderPath);
    console.log('=======================');

    // Show saving indicator
    const saveBtn = document.getElementById('save-btn');
    const originalText = saveBtn.innerHTML;
    saveBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Saving...';
    saveBtn.disabled = true;

    // Create the request body
    const requestBody = JSON.stringify({ 
        title: title || '',
        content: content || '',
        folder: folderPath || '',
        oldTitle: oldTitle || ''
    });
    
    console.log('9. Request body:', requestBody);
    console.log('10. Request body length:', requestBody.length);
    
    // Send request to save note
    fetch('/save', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: requestBody
    })
    .then(response => {
        console.log('11. Response status:', response.status);
        if (!response.ok) {
            return response.text().then(text => {
                console.log('12. Error response:', text);
                throw new Error(`Server returned ${response.status}: ${text}`);
            });
        }
        return response.json();
    })
    .then(data => {
        console.log('13. Success response:', data);
        isDirty = false;
        lastSavedContent = content;
        updateSaveButton();
        
        // Redirect to the view page
        if (data.redirect) {
            window.location.href = data.redirect;
        }
        
        // Show success message
        showNotification('Note saved successfully', 'success');
    })
    .catch(error => {
        console.error('14. Error caught:', error);
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