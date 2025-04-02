// View page specific JavaScript functions

// Function to confirm note deletion
function confirmDelete() {
    if (confirm('Are you sure you want to delete this note?')) {
        const loader = document.getElementById('loader');
        loader.classList.add('active');
        
        // Get the note title from the URL
        const pathParts = window.location.pathname.split('/');
        const title = pathParts[pathParts.length - 1];
        
        // Get the folder path from the URL query parameters
        const urlParams = new URLSearchParams(window.location.search);
        const folderPath = urlParams.get('folder');
        
        // Construct the delete URL with folder parameter if present
        let deleteUrl = `/delete/${title}`;
        if (folderPath) {
            deleteUrl += `?folder=${encodeURIComponent(folderPath)}`;
        }
        
        // Send delete request
        fetch(deleteUrl, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            }
        })
        .then(response => {
            if (!response.ok) {
                throw new Error('Failed to delete note');
            }
            return response.json();
        })
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
            loader.classList.remove('active');
        });
    }
}

// Initialize view page functionality
document.addEventListener('DOMContentLoaded', function() {
    // Add any additional initialization code here
}); 