import React, { useState, useEffect } from 'react';
import { createRoot } from 'react-dom/client';

const App: React.FC = () => {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState([]);
  const [error, setError] = useState('');
  const [page, setPage] = useState(0); // Current page
  const [size, setSize] = useState(10); // Results per page
  const [isSearching, setIsSearching] = useState(false); // Track if search is active

  const fetchResults = async () => {
    try {
      const endpoint = isSearching ? `/query?q=${query}` : '/all';
      const response = await fetch(`${endpoint}&f=${page * size}&s=${size}`);
      if (!response.ok) throw new Error('Failed to fetch');
      const data = await response.json();
      setResults(data.torrents);
    } catch (err) {
      setError((err as Error).message);
    }
  };

  useEffect(() => {
    fetchResults();
  }, [page]); // Fetch results when the page changes

  const handleSearch = () => {
    setIsSearching(true);
    setPage(0); // Reset to first page
    fetchResults();
  };

  const handleAll = () => {
    setIsSearching(false);
    setPage(0); // Reset to first page
    fetchResults();
  };

  const handleTorrent = async (hash: string) => {
    try {
      const response = await fetch(`/torrent?h=${hash}`);
      if (!response.ok) throw new Error('Failed to fetch');
      const data = await response.json();
      setResults(data.torrents);
    } catch (err) {
      setError((err as Error).message);
    }
  };

  return (
    <div>
      <h1>Welcome to Torrent Search</h1>
      <input
        type="text"
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        placeholder="Search for torrents"
      />
      <button onClick={handleSearch}>Search</button>
      <button onClick={handleAll}>Get All Torrents</button>
      {error && <p style={{ color: 'red' }}>{error}</p>}
      <ul>
        {results && results.map((torrent: any) => (
          <li key={torrent.infohashHex}>
            {torrent.name} - {torrent.length} bytes
            <button onClick={() => handleTorrent(torrent.infohashHex)}>Details</button>
          </li>
        ))}
      </ul>
      <div>
        <button onClick={() => setPage((prev) => Math.max(prev - 1, 0))} disabled={page === 0}>
          Previous
        </button>
        <span>Page {page + 1}</span>
        <button onClick={() => setPage((prev) => prev + 1)}>
          Next
        </button>
      </div>
    </div>
  );
};

const container = document.getElementById('root');
if (container) {
  const root = createRoot(container);
  root.render(<App />);
} else {
  console.error("Root container not found");
}
