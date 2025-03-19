tinymce.init({
    selector: '#editor',
    height: 500,
    plugins: [
        'advlist', 'autolink', 'lists', 'link', 'image', 'charmap', 'preview',
        'anchor', 'searchreplace', 'visualblocks', 'code', 'fullscreen',
        'insertdatetime', 'media', 'table', 'help', 'wordcount'
    ],
    toolbar: 'undo redo | blocks | ' +
        'bold italic backcolor | alignleft aligncenter ' +
        'alignright alignjustify | bullist numlist outdent indent | ' +
        'removeformat | help',
    content_style: `
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            font-size: 16px;
            line-height: 1.6;
            padding: 20px;
            margin: 0;
            background-color: #ffffff;
            color: #222f3e;
        }
        p { margin: 0 0 1em 0; }
        table { border-collapse: collapse; }
        table td, table th { border: 1px solid #ccc; padding: 8px; }
        a { color: #0066cc; }
    `,
    skin: 'oxide',
    content_css: false,
    base_url: '/static/js/tinymce',
    promotion: false,
    branding: false,
    setup: function(editor) {
        // Update editor theme when page theme changes
        const observer = new MutationObserver(function(mutations) {
            mutations.forEach(function(mutation) {
                if (mutation.attributeName === 'data-theme') {
                    editor.dom.setStyle(editor.getBody(), 'background-color', 
                        document.documentElement.getAttribute('data-theme') === 'dark' 
                            ? 'var(--bg-secondary)' 
                            : 'var(--bg-primary)'
                    );
                    editor.dom.setStyle(editor.getBody(), 'color',
                        'var(--text-primary)'
                    );
                }
            });
        });
        
        observer.observe(document.documentElement, {
            attributes: true,
            attributeFilter: ['data-theme']
        });
    }
}); 