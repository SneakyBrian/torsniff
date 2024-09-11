document.addEventListener('DOMContentLoaded', function() {
    console.log("Frontend is ready!");

    // Example of fetching data from the backend
    fetch('/all')
        .then(response => response.json())
        .then(data => {
            console.log(data);
            // You can now manipulate the DOM to display data
        })
        .catch(error => console.error('Error fetching data:', error));
});
