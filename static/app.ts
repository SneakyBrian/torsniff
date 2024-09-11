document.addEventListener('DOMContentLoaded', () => {
    console.log("Frontend is ready!");

    const form = document.getElementById('searchForm') as HTMLFormElement;
    const resultsList = document.getElementById('results') as HTMLUListElement;

    form.addEventListener('submit', (event: Event) => {
        event.preventDefault();
        const query = (document.getElementById('searchQuery') as HTMLInputElement).value;
        searchTorrents(query);
    });

    function searchTorrents(query: string): void {
        fetch(`/query?q=${encodeURIComponent(query)}`)
            .then(response => response.json())
            .then((data: { torrents: Torrent[] }) => {
                displayResults(data.torrents);
            })
            .catch(error => console.error('Error fetching data:', error));
    }

    function displayResults(torrents: Torrent[]): void {
        resultsList.innerHTML = '';
        if (torrents && Array.isArray(torrents) && torrents.length > 0) {
            torrents.forEach(torrent => {
                const li = document.createElement('li');
                li.textContent = `${torrent.name} (${torrent.length} bytes)`;
                resultsList.appendChild(li);
            });
        } else {
            const li = document.createElement('li');
            li.textContent = 'No torrents found.';
            resultsList.appendChild(li);
        }
    }

    // Fetch all torrents on page load
    fetch('/all')
        .then(response => response.json())
        .then((data: { torrents: Torrent[] }) => {
            displayResults(data.torrents);
        })
        .catch(error => console.error('Error fetching data:', error));
});

// Define the Torrent type
interface Torrent {
    name: string;
    length: number;
}
