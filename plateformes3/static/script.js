// script.js
document.addEventListener('DOMContentLoaded', function() {
    // Example: Adding a notification when a bucket is created or deleted
    document.addEventListener('htmx:afterRequest', function(evt) {
        if (evt.detail.requestConfig.verb === "POST") {
            alert("Bucket created successfully!");
        } else if (evt.detail.requestConfig.verb === "DELETE") {
            alert("Bucket deleted successfully!");
        }
    });
});
