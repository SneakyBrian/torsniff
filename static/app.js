document.addEventListener('DOMContentLoaded', function() {
    console.log("Frontend is ready!");

    const form = document.getElementById('searchForm');
    const resultsList = document.getElementById('results');

    form.addEventListener('submit', function(event) {
        event.preventDefault();
        const query = document.getElementById('searchQuery').value;
        searchTorrents(query);
    });

    function searchTorrents(query) {
        fetch(`/query?q=${encodeURIComponent(query)}`)
            .then(response => response.json())
            .then(data => {
                displayResults(data.torrents);
            })
            .catch(error => console.error('Error fetching data:', error));
    }

    function displayResults(torrents) {
        resultsList.innerHTML = '';
        if (Array.isArray(torrents)) {
            torrents.forEach(torrent => {
                const li = document.createElement('li');
                li.textContent = `${torrent.name} (${torrent.length} bytes)`;
                resultsList.appendChild(li);
            });
        } else {
            console.error('Invalid data format:', torrents);
        }
    }

    // Fetch all torrents on page load
    fetch('/all')
        .then(response => response.json())
        .then(data => {
            displayResults(data.torrents);
        })
        .catch(error => console.error('Error fetching data:', error));
});
