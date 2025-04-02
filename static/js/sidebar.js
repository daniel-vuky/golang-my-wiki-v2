// Sidebar functionality
function setupTreeExpandCollapse() {
    // Get all expand icons
    const expandIcons = document.querySelectorAll('.expand-icon');
    
    expandIcons.forEach(icon => {
        icon.addEventListener('click', handleExpandCollapse);
    });
}

// Sync functionality
function setupSyncButton() {
    const syncBtn = document.getElementById('sync-btn');
    if (!syncBtn) return;

    syncBtn.addEventListener('click', async () => {
        // Disable button and show syncing state
        syncBtn.disabled = true;
        syncBtn.classList.add('syncing');
        syncBtn.innerHTML = '<i class="fas fa-sync"></i> Syncing...';

        try {
            const response = await fetch('/api/sync', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                }
            });

            if (!response.ok) {
                throw new Error('Sync failed');
            }

            // Show success state
            syncBtn.innerHTML = '<i class="fas fa-check"></i> Sync Complete';
            setTimeout(() => {
                syncBtn.innerHTML = '<i class="fas fa-sync"></i> Sync with GitHub';
            }, 2000);

            // Reload the page to show updated content
            window.location.reload();
        } catch (error) {
            console.error('Sync error:', error);
            // Show error state
            syncBtn.innerHTML = '<i class="fas fa-exclamation-circle"></i> Sync Failed';
            setTimeout(() => {
                syncBtn.innerHTML = '<i class="fas fa-sync"></i> Sync with GitHub';
            }, 2000);
        } finally {
            // Re-enable button and remove syncing state
            syncBtn.disabled = false;
            syncBtn.classList.remove('syncing');
        }
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

function loadFolderChildren(folderPath, subtree, callback) {
    // Get folder children via API
    const apiUrl = folderPath === '' 
        ? '/api/folders/children/' 
        : `/api/folders/children/${folderPath}`;
    
    console.log("Loading children for folder:", folderPath);
    
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
                
                const isActive = child.Path === window.sidebarData.currentPath;
                
                if (child.IsNote) {
                    // Create note link - use view path for notes
                    let noteTitle = child.Name;
                    let noteFolderPath = child.Path.substring(0, child.Path.lastIndexOf('/')); // Extract folder path from note path
                    html += `<a href="/view/${noteTitle}?folder=${noteFolderPath}" class="tree-link ${isActive ? 'active' : ''}">
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

function expandPathToCurrentFolder(currentPath) {
    if (!currentPath) return;
    console.log("Expanding path to current folder:", currentPath);
    
    // Build all parent paths that need to be expanded
    const parts = currentPath.split('/');
    const parentPaths = [];
    
    let currentParentPath = '';
    for (let i = 0; i < parts.length; i++) {
        if (i > 0) currentParentPath += '/';
        currentParentPath += parts[i];
        parentPaths.push(currentParentPath);
    }
    
    console.log("Parent paths to expand:", parentPaths);
    
    // First mark the current path as active
    document.querySelectorAll('.tree-link.active').forEach(link => {
        link.classList.remove('active');
    });
    
    // Update current path as active
    const currentItem = document.querySelector(`.tree-item[data-path="${currentPath}"]`);
    if (currentItem) {
        const currentLink = currentItem.querySelector('.tree-link');
        if (currentLink) {
            currentLink.classList.add('active');
            // Update folder icon
            const icon = currentLink.querySelector('i.fa-folder');
            if (icon) icon.className = 'fas fa-folder-open';
        }
    }
    
    // Process parents from root to current path (important for proper loading sequence)
    for (let i = 0; i < parentPaths.length; i++) {
        const path = parentPaths[i];
        console.log("Processing path:", path);
        
        // Skip the current path - we only want to expand parents
        if (i === parentPaths.length - 1 && path === currentPath) {
            console.log("Skipping current path:", path);
            continue;
        }
        
        // Find the parent item
        const item = document.querySelector(`.tree-item[data-path="${path}"]`);
        if (!item) {
            console.log("Item not found:", path);
            continue;
        }
        
        console.log("Found item:", path, "Has children:", item.classList.contains('has-children'));
        
        // Only process items with children that need expansion
        if (item.classList.contains('has-children') && !item.classList.contains('expanded')) {
            console.log("Expanding:", path);
            
            // Mark as expanded
            item.classList.add('expanded');
            
            // Find the subtree
            const subtree = item.querySelector('.subtree');
            if (!subtree) {
                console.log("No subtree found for:", path);
                continue;
            }
            
            // Load children if needed
            if (subtree.classList.contains('not-loaded')) {
                console.log("Loading children for:", path);
                loadFolderChildren(path, subtree);
            }
            
            // Show the subtree
            subtree.style.display = 'block';
            
            // Rotate the expand icon
            const expandIcon = item.querySelector('.expand-icon i');
            if (expandIcon) {
                expandIcon.style.transform = 'rotate(90deg)';
            }
        }
    }
}

function highlightCurrentNote(noteTitle, folderPath) {
    if (!noteTitle || !folderPath) return;
    
    console.log("Trying to highlight note:", noteTitle, "in folder:", folderPath);
    
    // Look for notes with matching title in the folder's children
    const folderItem = document.querySelector(`.tree-item[data-path="${folderPath}"]`);
    if (folderItem) {
        const subtree = folderItem.querySelector('.subtree');
        if (subtree && !subtree.classList.contains('not-loaded')) {
            // Folder's children are already loaded, try to find the note
            const noteItems = subtree.querySelectorAll(`.tree-item[data-type="note"]`);
            let foundNote = false;
            
            for (const noteItem of noteItems) {
                const noteLink = noteItem.querySelector('.tree-link');
                if (noteLink) {
                    // Trim whitespace and compare with note title
                    const linkText = noteLink.textContent.trim();
                    console.log("Comparing:", linkText, "with:", noteTitle);
                    
                    if (linkText === noteTitle) {
                        noteLink.classList.add('active');
                        console.log("Found and highlighted note:", noteTitle);
                        foundNote = true;
                        break;
                    }
                }
            }
            
            if (!foundNote) {
                console.log("Note not found in loaded children. Note title:", noteTitle);
            }
        } else {
            // Need to load the folder's children
            console.log("Loading folder children to find note");
            loadFolderChildren(folderPath, subtree, function() {
                // After loading, try again to find the note
                const noteItems = subtree.querySelectorAll(`.tree-item[data-type="note"]`);
                for (const noteItem of noteItems) {
                    const noteLink = noteItem.querySelector('.tree-link');
                    if (noteLink && noteLink.textContent.trim() === noteTitle) {
                        noteLink.classList.add('active');
                        console.log("Found and highlighted note after loading:", noteTitle);
                        return;
                    }
                }
            });
        }
    } else {
        console.log("Folder item not found for path:", folderPath);
    }
}

// Initialize sidebar functionality
document.addEventListener('DOMContentLoaded', function() {
    // Initialize tree expand/collapse
    setupTreeExpandCollapse();
    
    // Initialize sync button
    setupSyncButton();
    
    // Auto-expand path to current folder/note
    const folderPath = window.sidebarData.folderPath;
    const noteTitle = window.sidebarData.noteTitle;
    
    if (folderPath) {
        console.log("Current note:", noteTitle, "in folder:", folderPath);
        
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
            
            console.log(`Processing folder at index ${index}: ${currentPath}`);
            
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
                console.log(`Folder ${currentPath} not found in sidebar`);
                // Folder not found in sidebar, try the next one
                processNextFolder(index + 1);
            }
        };
        
        // Start processing from the first folder
        processNextFolder(0);
    }
}); 