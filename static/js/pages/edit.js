let editor;
let isDirty = false;
let lastSavedContent = '';

document.addEventListener('DOMContentLoaded', function() {
    initEditor();
    initEventListeners();
});

function isDarkTheme() {
    return document.body.classList.contains('dark-theme');
}

function syncHljsTheme() {
    const dark = isDarkTheme();
    const light = document.getElementById('hljs-light');
    const darkSheet = document.getElementById('hljs-dark');
    if (light) light.disabled = dark;
    if (darkSheet) darkSheet.disabled = !dark;
}

// Dark theme is handled entirely by CSS (body.dark-theme .toastui-*)
// This is a no-op kept for the toggle hook structure below
function applyEditorTheme() {}

function initEditor() {
    const initialContent = document.getElementById('raw-content').value;

    // Safely resolve the code syntax highlight plugin
    let plugins = [];
    try {
        const csh = toastui.Editor.plugin.codeSyntaxHighlight;
        if (csh) plugins = [csh];
    } catch (e) {
        console.warn('codeSyntaxHighlight plugin not available:', e);
    }

    editor = new toastui.Editor({
        el: document.getElementById('editor'),
        height: '550px',
        initialEditType: 'wysiwyg',
        previewStyle: 'vertical',
        initialValue: initialContent,
        theme: 'light',
        plugins: plugins,
        // Use default toolbar so all buttons (including codeblock) are available
        events: {
            change: function() {
                isDirty = editor.getMarkdown() !== lastSavedContent;
            }
        }
    });

    lastSavedContent = editor.getMarkdown();

    // Apply theme immediately after render
    applyEditorTheme();
    syncHljsTheme();

    // Hook into theme toggle
    const _orig = window.toggleTheme;
    window.toggleTheme = function() {
        _orig();
        applyEditorTheme();
        syncHljsTheme();
    };
}

function initEventListeners() {
    const form = document.getElementById('note-form');
    const saveBtn = document.getElementById('save-btn');

    if (form) form.addEventListener('submit', function(e) { e.preventDefault(); saveContent(); });
    if (saveBtn) saveBtn.addEventListener('click', function(e) { e.preventDefault(); saveContent(); });

    document.addEventListener('keydown', function(e) {
        if ((e.ctrlKey || e.metaKey) && e.key === 's') { e.preventDefault(); saveContent(); }
    });
}

function saveContent() {
    const titleInput = document.getElementById('title');
    const folderPathInput = document.querySelector('input[name="folder_path"]');
    const originalTitleInput = document.querySelector('input[name="original_title"]');

    if (!editor || !titleInput || !folderPathInput) {
        console.error('Required elements not found');
        return;
    }

    const content = editor.getMarkdown();
    const title = titleInput.value;
    const folderPath = folderPathInput.value;
    const oldTitle = originalTitleInput ? originalTitleInput.value : '';

    const saveBtn = document.getElementById('save-btn');
    const originalText = saveBtn.innerHTML;
    saveBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Saving...';
    saveBtn.disabled = true;

    fetch('/save', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ title, content, folder: folderPath, oldTitle })
    })
    .then(function(response) {
        if (!response.ok) {
            return response.text().then(function(text) {
                throw new Error('Server returned ' + response.status + ': ' + text);
            });
        }
        return response.json();
    })
    .then(function(data) {
        isDirty = false;
        lastSavedContent = content;
        if (data.redirect) window.location.href = data.redirect;
        showNotification('Note saved successfully', 'success');
    })
    .catch(function(error) {
        showNotification('Error saving note: ' + error.message, 'error');
    })
    .finally(function() {
        saveBtn.innerHTML = originalText;
        saveBtn.disabled = false;
    });
}

function showNotification(message, type) {
    type = type || 'info';
    const n = document.createElement('div');
    n.className = 'notification ' + type;
    n.innerHTML = '<i class="fas ' + (type === 'success' ? 'fa-check-circle' : 'fa-exclamation-circle') + '"></i> ' + message;
    document.body.appendChild(n);
    setTimeout(function() { n.remove(); }, 3000);
}

window.addEventListener('beforeunload', function(e) {
    if (isDirty) { e.preventDefault(); e.returnValue = ''; }
});
