// Folder tree expand/collapse functionality
document.addEventListener('DOMContentLoaded', function() {
    // Initialize tree expand/collapse
    setupTreeExpandCollapse();
    
    // Auto-expand path to current folder
    const currentPath = document.querySelector('meta[name="current-path"]')?.content;
    if (currentPath) {
        expandPathToCurrentFolder(currentPath);
    }
    
    // Refresh button functionality
    const refreshButton = document.getElementById('refreshButton');
    if (refreshButton) {
        refreshButton.addEventListener('click', function(e) {
            e.preventDefault();
            
            // Show loading state
            this.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Refreshing...';
            this.classList.add('loading');
            
            // Redirect to refresh URL
            window.location.href = this.getAttribute('href');
        });
    }
    
    // Subcategory popup functionality
    const subcategoryBtn = document.getElementById('btn-subcategory');
    const subcategoryPopup = document.getElementById('subcategory-popup');
    const cancelBtn = document.getElementById('cancel-subcategory');
    const saveBtn = document.getElementById('save-subcategory');
    const subcategoryInput = document.getElementById('subcategory-name');
    
    subcategoryBtn.addEventListener('click', function() {
        subcategoryPopup.classList.add('active');
        subcategoryInput.focus();
    });
    
    cancelBtn.addEventListener('click', function() {
        subcategoryPopup.classList.remove('active');
        subcategoryInput.value = '';
    });
    
    subcategoryPopup.addEventListener('click', function(e) {
        if (e.target === subcategoryPopup) {
            subcategoryPopup.classList.remove('active');
            subcategoryInput.value = '';
        }
    });
    
    saveBtn.addEventListener('click', createSubcategory);
    subcategoryInput.addEventListener('keypress', function(e) {
        if (e.key === 'Enter') {
            createSubcategory();
        }
    });
});

function createSubcategory() {
    const subcategoryName = document.getElementById('subcategory-name').value.trim();
    const currentFolderPath = window.sidebarData.folderPath;
    const parentFolderSha = window.sidebarData.parentFolderSha;
    
    if (!subcategoryName) {
        alert('Please enter a subcategory name');
        return;
    }
    
    // Create the request body with the correct parent path and SHA
    const requestBody = JSON.stringify({ 
        name: subcategoryName,
        parentPath: currentFolderPath,
        sha: parentFolderSha
    });
    
    // Show loading overlay
    const loadingOverlay = document.getElementById('loading-overlay');
    loadingOverlay.classList.add('active');
    
    // Send request to create subcategory
    fetch('/category/create', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: requestBody
    })
    .then(response => response.json())
    .then(data => {
        if (data.success) {
            // Success - reload the page
            window.location.reload();
        } else {
            // Error - show message
            throw new Error(data.error || 'Failed to create subcategory');
        }
    })
    .catch(error => {
        console.error('Error:', error);
        alert(error.message || 'Error creating subcategory. Please try again.');
    })
    .finally(() => {
        // Hide loading overlay
        loadingOverlay.classList.remove('active');
        
        // Reset form and close popup
        document.getElementById('subcategory-name').value = '';
        document.getElementById('subcategory-popup').classList.remove('active');
    });
}

// Delete folder functionality
let folderToDelete = null;
const deleteFolderPopup = document.getElementById('delete-folder-popup');
const cancelDeleteFolderBtn = document.getElementById('cancel-delete-folder');
const confirmDeleteFolderBtn = document.getElementById('confirm-delete-folder');

function confirmDeleteFolder(folderPath) {
    folderToDelete = folderPath;
    deleteFolderPopup.classList.add('active');
}

if (cancelDeleteFolderBtn) {
    cancelDeleteFolderBtn.addEventListener('click', () => {
        deleteFolderPopup.classList.remove('active');
        folderToDelete = null;
    });
}

if (confirmDeleteFolderBtn) {
    confirmDeleteFolderBtn.addEventListener('click', async () => {
        if (!folderToDelete) return;
        
        // Show loading overlay
        const loadingOverlay = document.getElementById('loading-overlay');
        const loadingText = loadingOverlay.querySelector('.loading-text');
        loadingText.textContent = 'Deleting...';
        loadingOverlay.classList.add('active');
        
        try {
            const response = await fetch(`/api/folder/delete?path=${encodeURIComponent(folderToDelete)}`, {
                method: 'DELETE',
            });
            
            const data = await response.json();
            if (response.ok) {
                // Hide loading overlay before redirect
                loadingOverlay.classList.remove('active');
                loadingText.textContent = 'Loading...';
                deleteFolderPopup.classList.remove('active');
                folderToDelete = null;
                
                // Redirect to the URL provided by the server
                window.location.replace(data.redirect);
            } else {
                throw new Error(data.error || 'Failed to delete folder');
            }
        } catch (error) {
            console.error('Error deleting folder:', error);
            alert(error.message || 'An error occurred while deleting the folder. Please try again.');
            // Hide loading overlay on error
            loadingOverlay.classList.remove('active');
            loadingText.textContent = 'Loading...';
            deleteFolderPopup.classList.remove('active');
            folderToDelete = null;
        }
    });
}

// Function to confirm note deletion
function confirmDelete(path) {
    if (confirm('Are you sure you want to delete this note?')) {
        // Show loading overlay
        const loadingOverlay = document.getElementById('loading-overlay');
        const loadingText = loadingOverlay.querySelector('.loading-text');
        loadingText.textContent = 'Deleting note...';
        loadingOverlay.classList.add('active');
        
        // Extract note title and folder path from the full path
        const pathParts = path.split('/');
        const noteTitle = pathParts[pathParts.length - 1];
        const folderPath = pathParts.slice(0, -1).join('/');
        
        // Send delete request
        fetch(`/delete/${noteTitle}?folder=${encodeURIComponent(folderPath)}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            }
        })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                // Redirect to the URL provided by the server
                window.location.href = data.redirect;
            } else {
                throw new Error(data.error || 'Failed to delete note');
            }
        })
        .catch(error => {
            console.error('Error:', error);
            alert(error.message || 'Error deleting note. Please try again.');
        })
        .finally(() => {
            // Hide loading overlay
            loadingOverlay.classList.remove('active');
            loadingText.textContent = 'Loading...';
        });
    }
} 