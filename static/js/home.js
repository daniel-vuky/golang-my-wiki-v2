// Home page specific JavaScript functions

// Function to confirm deletion
function confirmDelete(title) {
    if (confirm('Are you sure you want to delete "' + title + '"?')) {
        fetch(`/delete/${title}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            }
        })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                window.location.reload();
            } else {
                throw new Error(data.error || 'Failed to delete item');
            }
        })
        .catch(error => {
            console.error('Error:', error);
            alert(error.message || 'Error deleting item. Please try again.');
        });
    }
}

// Function to create a new category
function createCategory() {
    const categoryName = document.getElementById('categoryName').value.trim();
    const saveBtn = document.getElementById('saveCategory');
    const categoryPopup = document.getElementById('categoryPopup');
    const categoryInput = document.getElementById('categoryName');
    
    if (!categoryName) {
        alert('Please enter a category name');
        return;
    }
    
    // Create the request body
    const requestBody = JSON.stringify({ name: categoryName });
    console.log('Sending request with body:', requestBody);
    
    // Show loading indicator
    saveBtn.disabled = true;
    saveBtn.innerHTML = 'Creating...';
    
    // Send request to create category
    fetch('/category/create', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: requestBody
    })
    .then(response => {
        console.log('Response status:', response.status);
        
        // Try to parse as JSON
        return response.text().then(text => {
            console.log('Response text:', text);
            
            try {
                // If it's valid JSON, parse it
                if (text && text.trim()) {
                    return JSON.parse(text);
                }
                return { success: response.ok };
            } catch (e) {
                console.error('JSON parse error:', e);
                // If it's not valid JSON, return the text
                return { 
                    success: false, 
                    error: 'Invalid server response: ' + text
                };
            }
        });
    })
    .then(data => {
        if (data.success) {
            // Success - reload the page
            window.location.reload();
        } else {
            // Error - show message
            throw new Error(data.error || 'Failed to create category');
        }
    })
    .catch(error => {
        console.error('Error:', error);
        alert(error.message || 'Error creating category. Please try again.');
    })
    .finally(() => {
        // Reset button state
        saveBtn.disabled = false;
        saveBtn.innerHTML = 'Save';
        
        // Close popup on error
        categoryPopup.classList.remove('active');
        categoryInput.value = '';
    });
}

// Initialize category creation functionality
document.addEventListener('DOMContentLoaded', function() {
    const newCategoryBtn = document.getElementById('newCategoryBtn');
    const categoryPopup = document.getElementById('categoryPopup');
    const cancelBtn = document.getElementById('cancelCategory');
    const saveBtn = document.getElementById('saveCategory');
    const categoryInput = document.getElementById('categoryName');
    
    if (newCategoryBtn && categoryPopup) {
        // Open popup on button click
        newCategoryBtn.addEventListener('click', function() {
            categoryPopup.classList.add('active');
            categoryInput.focus();
        });
        
        // Close popup on cancel button click
        cancelBtn.addEventListener('click', function() {
            categoryPopup.classList.remove('active');
            categoryInput.value = '';
        });
        
        // Close popup when clicking outside
        categoryPopup.addEventListener('click', function(e) {
            if (e.target === categoryPopup) {
                categoryPopup.classList.remove('active');
                categoryInput.value = '';
            }
        });
        
        // Save category on button click
        saveBtn.addEventListener('click', createCategory);
        
        // Save category on Enter key
        categoryInput.addEventListener('keypress', function(e) {
            if (e.key === 'Enter') {
                createCategory();
            }
        });
    }
}); 