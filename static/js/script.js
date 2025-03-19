document.addEventListener('DOMContentLoaded', function() {
    // Set active menu item based on the current URL
    const currentPath = window.location.pathname;
    const navLinks = document.querySelectorAll('.sidebar-nav ul li a');
    
    navLinks.forEach(link => {
        const href = link.getAttribute('href');
        
        // For homepage
        if (currentPath === '/' && href === '/') {
            link.parentElement.classList.add('active');
        }
        // For view pages
        else if (currentPath.includes('/view/') && href.includes('/view/')) {
            const currentTitle = currentPath.split('/view/')[1];
            const linkTitle = href.split('/view/')[1];
            
            if (currentTitle === linkTitle) {
                link.parentElement.classList.add('active');
            } else {
                link.parentElement.classList.remove('active');
            }
        }
        // For edit pages, keep the corresponding view page active
        else if (currentPath.includes('/edit/') && href.includes('/view/')) {
            const currentTitle = currentPath.split('/edit/')[1];
            const linkTitle = href.split('/view/')[1];
            
            if (currentTitle === linkTitle) {
                link.parentElement.classList.add('active');
            } else {
                link.parentElement.classList.remove('active');
            }
        }
        // For other pages, remove active class
        else {
            link.parentElement.classList.remove('active');
        }
    });
    
    // Add tab support in the editor
    const contentEditor = document.getElementById('content-editor');
    if (contentEditor) {
        contentEditor.addEventListener('keydown', function(e) {
            if (e.key === 'Tab') {
                e.preventDefault();
                
                // Insert tab at cursor position
                const start = this.selectionStart;
                const end = this.selectionEnd;
                
                this.value = this.value.substring(0, start) + '    ' + this.value.substring(end);
                this.selectionStart = this.selectionEnd = start + 4;
            }
        });
    }

    // Theme toggling functionality
    const themeToggle = document.getElementById('theme-toggle');
    const themeIcon = themeToggle.querySelector('i');
    
    // Check for saved theme preference or respect OS theme preference
    const savedTheme = localStorage.getItem('theme');
    const prefersDarkScheme = window.matchMedia('(prefers-color-scheme: dark)');
    
    // Set initial theme
    if (savedTheme === 'dark' || (!savedTheme && prefersDarkScheme.matches)) {
        document.body.classList.add('dark-theme');
        document.documentElement.setAttribute('data-theme', 'dark');
        themeIcon.classList.remove('fa-moon');
        themeIcon.classList.add('fa-sun');
    }
    
    // Toggle theme when button is clicked
    themeToggle.addEventListener('click', () => {
        const isDarkTheme = document.body.classList.contains('dark-theme');
        
        if (isDarkTheme) {
            // Switch to light theme
            document.body.classList.remove('dark-theme');
            document.documentElement.setAttribute('data-theme', 'light');
            localStorage.setItem('theme', 'light');
            themeIcon.classList.remove('fa-sun');
            themeIcon.classList.add('fa-moon');
        } else {
            // Switch to dark theme
            document.body.classList.add('dark-theme');
            document.documentElement.setAttribute('data-theme', 'dark');
            localStorage.setItem('theme', 'dark');
            themeIcon.classList.remove('fa-moon');
            themeIcon.classList.add('fa-sun');
        }
    });
}); 