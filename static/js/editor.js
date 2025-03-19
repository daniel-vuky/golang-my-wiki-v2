document.addEventListener('DOMContentLoaded', function() {
    const editor = document.getElementById('content-editor');
    const form = editor.closest('form');
    const hiddenInput = document.getElementById('editor-content');

    // Initialize contenteditable
    editor.contentEditable = true;
    editor.designMode = 'on';

    // Handle form submission
    form.addEventListener('submit', function(e) {
        e.preventDefault();
        // Get the HTML content and set it to the hidden input
        hiddenInput.value = editor.innerHTML;
        form.submit();
    });

    // Handle toolbar buttons
    window.execCommand = function(command, value = null) {
        document.execCommand(command, false, value);
        editor.focus();
    };

    // Set initial content if exists
    if (hiddenInput.value) {
        // Create a temporary div to parse the HTML
        const temp = document.createElement('div');
        temp.innerHTML = hiddenInput.value;
        // Clean up empty divs
        const cleanContent = temp.innerHTML.replace(/<div><br><\/div>/g, '<br>');
        editor.innerHTML = cleanContent;
    }

    // Prevent default drag and drop behavior
    editor.addEventListener('dragstart', function(e) {
        e.preventDefault();
    });

    // Handle paste to clean up content
    editor.addEventListener('paste', function(e) {
        e.preventDefault();
        const text = e.clipboardData.getData('text/plain');
        document.execCommand('insertText', false, text);
    });
}); 