function setupTreeExpandCollapse() {
    // Get all expand icons
    const expandIcons = document.querySelectorAll('.expand-icon');
    
    expandIcons.forEach(icon => {
        icon.addEventListener('click', handleExpandCollapse);
    });
}

function handleExpandCollapse(event) {
    event.stopPropagation(); // Prevent bubbling
    
    const treeItem = this.closest('.tree-item');
    const subtree = treeItem.querySelector('.subtree');
    
    if (!treeItem.classList.contains('expanded')) {
        // Expanding - Check if we need to load children
        if (subtree && subtree.classList.contains('not-loaded')) {
            const path = treeItem.getAttribute('data-path');
            loadFolderChildren(path, subtree);
        }
    }
    
    // Toggle expanded state
    treeItem.classList.toggle('expanded');
    
    // Toggle visibility of the subtree
    if (subtree) {
        if (treeItem.classList.contains('expanded')) {
            subtree.style.display = 'block';
        } else {
            subtree.style.display = 'none';
        }
    }
}

// Function to load folder children from API
function loadFolderChildren(folderPath, subtree, callback) {
    // Get folder children via API
    const apiUrl = folderPath === '' 
        ? '/api/folders/children/' 
        : `/api/folders/children/${folderPath}`;
    
    fetch(apiUrl)
        .then(response => response.json())
        .then(data => {
            // Clear loading indicator
            subtree.innerHTML = '';
            subtree.classList.remove('not-loaded');
            
            // No children case
            if (data.children.length === 0) {
                subtree.innerHTML = '<li class="empty-item">No subfolders</li>';
                if (callback) callback();
                return;
            }
            
            const folderPath = document.querySelector('input[name="folder_path"]')?.value || '';
            const noteTitle = document.querySelector('input[name="original_title"]')?.value || '';
            
            // Add children to the subtree
            data.children.forEach(child => {
                const childItem = document.createElement('li');
                childItem.className = `tree-item ${child.HasChildren ? 'has-children' : ''}`;
                childItem.setAttribute('data-path', child.Path);
                childItem.setAttribute('data-type', child.IsNote ? 'note' : 'folder');
                
                let html = '';
                if (child.HasChildren && !child.IsNote) {
                    html += `<span class="expand-icon"><i class="fas fa-caret-right"></i></span>`;
                }
                
                const isActive = child.Path === document.querySelector('input[name="current_path"]')?.value;
                // Check if this child is the current note we're viewing
                const isCurrentNote = child.IsNote && child.Name === noteTitle && 
                                     folderPath && (child.Path === folderPath + '/' + noteTitle || 
                                                   child.Path.endsWith('/' + noteTitle));
                
                if (child.IsNote) {
                    // Create note link - use view path for notes
                    let noteTitle = child.Name;
                    let noteFolderPath = child.Path.substring(0, child.Path.lastIndexOf('/')); // Extract folder path from note path
                    html += `<a href="/view/${noteTitle}?folder=${noteFolderPath}" class="tree-link ${isCurrentNote ? 'active' : ''}">
                        <i class="fas fa-file-alt"></i> ${child.Name}
                    </a>`;
                } else {
                    // Create folder link
                    html += `<a href="/category/${child.Path}" class="tree-link ${isActive ? 'active' : ''}">
                        <i class="fas fa-folder${isActive ? '-open' : ''}"></i> ${child.Name}
                    </a>`;
                    
                    if (child.HasChildren) {
                        html += `<ul class="subtree not-loaded"></ul>`;
                    }
                }
                
                childItem.innerHTML = html;
                subtree.appendChild(childItem);
                
                // Add event listener to new expand icon
                const expandIcon = childItem.querySelector('.expand-icon');
                if (expandIcon) {
                    expandIcon.addEventListener('click', handleExpandCollapse);
                }
                
                // If this is the current note, highlight it
                if (isCurrentNote) {
                    const noteLink = childItem.querySelector('.tree-link');
                    if (noteLink) noteLink.classList.add('active');
                }
            });
            
            // After fully loading children, call the callback if provided
            if (callback) callback();
        })
        .catch(error => {
            console.error('Error loading folder children:', error);
            subtree.innerHTML = '<li class="error-item">Error loading folders</li>';
            if (callback) callback();
        });
}

// Function to highlight the current note in the tree
function highlightCurrentNote(noteTitle, folderPath) {
    if (!noteTitle || !folderPath) return;
    
    // Look for notes with matching title in the folder's children
    const folderItem = document.querySelector(`.tree-item[data-path="${folderPath}"]`);
    if (folderItem) {
        const subtree = folderItem.querySelector('.subtree');
        if (subtree && !subtree.classList.contains('not-loaded')) {
            // Folder's children are already loaded, try to find the note
            const noteItems = subtree.querySelectorAll(`.tree-item[data-type="note"]`);
            for (const noteItem of noteItems) {
                const noteLink = noteItem.querySelector('.tree-link');
                if (noteLink && noteLink.textContent.trim() === noteTitle) {
                    noteLink.classList.add('active');
                    return;
                }
            }
        } else if (subtree) {
            // Need to load the folder's children
            loadFolderChildren(folderPath, subtree, function() {
                // After loading, try again to find the note
                const noteItems = subtree.querySelectorAll(`.tree-item[data-type="note"]`);
                for (const noteItem of noteItems) {
                    const noteLink = noteItem.querySelector('.tree-link');
                    if (noteLink && noteLink.textContent.trim() === noteTitle) {
                        noteLink.classList.add('active');
                        return;
                    }
                }
            });
        }
    }
}

// Initialize sidebar when DOM is loaded
document.addEventListener('DOMContentLoaded', function() {
    // Initialize tree expand/collapse
    setupTreeExpandCollapse();
    
    // Auto-expand path to current folder/note
    const folderPath = document.querySelector('input[name="folder_path"]')?.value;
    const noteTitle = document.querySelector('input[name="original_title"]')?.value;
    
    if (folderPath) {
        // Force expand all parent folders first 
        const parts = folderPath.split('/');
        let currentPath = '';
        
        // Process each folder in the path sequentially
        const processNextFolder = (index) => {
            if (index >= parts.length) {
                // We've processed all folders in the path, now highlight the note
                setTimeout(() => {
                    highlightCurrentNote(noteTitle, folderPath);
                }, 300);
                return;
            }
            
            // Build the path incrementally
            if (index > 0) currentPath += '/';
            currentPath += parts[index];
            
            // Find this folder in the tree
            const folderItem = document.querySelector(`.tree-item[data-path="${currentPath}"]`);
            if (folderItem) {
                // Mark as expanded
                folderItem.classList.add('expanded');
                
                // Find the subtree
                const subtree = folderItem.querySelector('.subtree');
                if (subtree) {
                    // Show the subtree
                    subtree.style.display = 'block';
                    
                    // Rotate the expand icon
                    const expandIcon = folderItem.querySelector('.expand-icon i');
                    if (expandIcon) {
                        expandIcon.style.transform = 'rotate(90deg)';
                    }
                    
                    // If not loaded yet, load its children before continuing
                    if (subtree.classList.contains('not-loaded')) {
                        loadFolderChildren(currentPath, subtree, () => {
                            // Process next folder in path after this one is loaded
                            processNextFolder(index + 1);
                        });
                    } else {
                        // Already loaded, continue to next folder
                        processNextFolder(index + 1);
                    }
                } else {
                    // No subtree, continue to next folder
                    processNextFolder(index + 1);
                }
            } else {
                // Folder not found in sidebar, try the next one
                processNextFolder(index + 1);
            }
        };
        
        // Start processing from the first folder
        processNextFolder(0);
    }
}); 