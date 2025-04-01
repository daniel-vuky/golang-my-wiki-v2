// Initialize sidebar data
window.sidebarData = {
    currentPath: "",
    folderPath: "",
    noteTitle: "",
    isNewPage: false
};

// Initialize sidebar functionality
function initializeSidebar() {
    setupTreeExpandCollapse();
    
    // Expand to current folder and highlight current note
    if (window.sidebarData.currentPath) {
        expandPathToCurrentFolder(window.sidebarData.currentPath);
    }
    
    // Highlight current note if not a new page
    if (!window.sidebarData.isNewPage && window.sidebarData.noteTitle && window.sidebarData.folderPath) {
        highlightCurrentNote(window.sidebarData.noteTitle, window.sidebarData.folderPath);
    }
}

// Wait for DOM and sidebar to be ready
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initializeSidebar);
} else {
    initializeSidebar();
} 