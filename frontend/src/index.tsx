import React, { useState } from 'react';
import { createRoot } from 'react-dom/client';

const App: React.FC = () => {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState([]);
  const [error, setError] = useState('');

  const handleSearch = async () => {
    try {
      const response = await fetch(`/query?q=${query}`);
      if (!response.ok) throw new Error('Failed to fetch');
      const data = await response.json();
      setResults(data.torrents);
    } catch (err) {
      setError((err as Error).message);
    }
  };

  const handleAll = async () => {
    try {
      const response = await fetch('/all');
      if (!response.ok) throw new Error('Failed to fetch');
      const data = await response.json();
      setResults(data.torrents);
    } catch (err) {
      setError((err as Error).message);
    }
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
        {results.map((torrent: any) => (
          <li key={torrent.InfohashHex}>
            {torrent.Name} - {torrent.Length} bytes
            <button onClick={() => handleTorrent(torrent.InfohashHex)}>Details</button>
          </li>
        ))}
      </ul>
    </div>
  );
};

const container = document.getElementById('root');
if (container) {
  const root = createRoot(container);
  root.render(<App />);
}
