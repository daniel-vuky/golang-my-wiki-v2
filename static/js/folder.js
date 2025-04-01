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
    const currentFolderPath = document.querySelector('meta[name="folder-path"]')?.content || '';
    const currentFolderSha = document.querySelector('meta[name="folder-sha"]')?.content || '';
    
    if (!subcategoryName) {
        alert('Please enter a subcategory name');
        return;
    }
    
    // Create the request body with the correct parent path and SHA
    const requestBody = JSON.stringify({ 
        name: subcategoryName,
        parentPath: currentFolderPath,
        sha: currentFolderSha
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