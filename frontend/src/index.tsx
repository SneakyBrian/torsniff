import React, { useState, useEffect } from 'react';
import 'bootstrap/dist/css/bootstrap.min.css';
import { createRoot } from 'react-dom/client';

const App: React.FC = () => {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState([]);
  const [error, setError] = useState('');
  const [page, setPage] = useState(0); // Current page
  const [size, setSize] = useState(10); // Results per page
  const [isSearching, setIsSearching] = useState(false); // Track if search is active
  const [selectedTorrent, setSelectedTorrent] = useState<any | null>(null); // State for selected torrent

  const fetchResults = async () => {
    try {
      const endpoint = isSearching ? `/query?q=${query}` : '/all?';
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
      setSelectedTorrent(data.torrents[0]); // Assuming the response contains a single torrent
    } catch (err) {
      setError((err as Error).message);
    }
  };

  // Utility function to format bytes
  const formatBytes = (bytes: number, decimals = 2) => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const dm = decimals < 0 ? 0 : decimals;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
  };

  const renderFiles = (files: any[]) => {
    const fileTree: any = {};

    files.forEach((file) => {
      const parts = file.name.split('/');
      let current = fileTree;

      parts.forEach((part: string, index: number) => {
        if (!current[part]) {
          current[part] = index === parts.length - 1 ? file.length : {};
        }
        current = current[part];
      });
    });

    const renderTree = (node: any, path: string[] = []) => {
      return Object.entries(node).map(([key, value]) => {
        const currentPath = [...path, key];
        if (typeof value === 'number') {
          return (
            <li key={currentPath.join('/')}>
              {key} - {formatBytes(value)} {/* Use the formatBytes function here */}
            </li>
          );
        }
        return (
          <li key={currentPath.join('/')}>
            <strong>{key}</strong>
            <ul>{renderTree(value, currentPath)}</ul>
          </li>
        );
      });
    };

    return <ul>{renderTree(fileTree)}</ul>;
  };

  return (
    <div className="container mt-5">
      <h1 className="text-center mb-4">Welcome to Torrent Search</h1>
      <div className="input-group mb-3">
        <input
          type="text"
          className="form-control"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Search for torrents"
        />
        <button className="btn btn-primary" onClick={handleSearch}>Search</button>
        <button className="btn btn-secondary" onClick={handleAll}>Get All Torrents</button>
      </div>
      {error && <p className="text-danger">{error}</p>}
      <ul className="list-group mb-3">
        {results && results.map((torrent: any) => (
          <li key={torrent.infohashHex} className="list-group-item d-flex justify-content-between align-items-center">
            {torrent.name} - {formatBytes(torrent.length)}
            <button className="btn btn-link" onClick={() => handleTorrent(torrent.infohashHex)}>Details</button>
          </li>
        ))}
      </ul>
      <div className="d-flex justify-content-between">
        <button className="btn btn-outline-primary" onClick={() => setPage((prev) => Math.max(prev - 1, 0))} disabled={page === 0}>
          Previous
        </button>
        <span>Page {page + 1}</span>
        <button className="btn btn-outline-primary" onClick={() => setPage((prev) => prev + 1)}>
          Next
        </button>
      </div>
      {selectedTorrent && (
        <div className="mt-4">
          <h2>Torrent Details</h2>
          <p>Name: {selectedTorrent.name}</p>
          <p>Size: {formatBytes(selectedTorrent.length)}</p>
          <h3>Files:</h3>
          {renderFiles(selectedTorrent.files)}
        </div>
      )}
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
