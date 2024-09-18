import React, { useState, useEffect, Suspense } from 'react';
import 'bootstrap/dist/css/bootstrap.min.css';
import { createRoot } from 'react-dom/client';

const App: React.FC = () => {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState([]);
  const [error, setError] = useState('');
  const [page, setPage] = useState(0);
  const [size, setSize] = useState(10);
  const [totalCount, setTotalCount] = useState<number | null>(null);
  const [isSearching, setIsSearching] = useState(false);
  const [hasMoreResults, setHasMoreResults] = useState(true);
  const [selectedTorrent, setSelectedTorrent] = useState<any | null>(null);
  const [trackers, setTrackers] = useState<string[]>([]);
  const [showModal, setShowModal] = useState(false);
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [torrentToDelete, setTorrentToDelete] = useState<string | null>(null);

  const fetchResults = async () => {
    try {
      const endpoint = isSearching ? `/query?q=${query}` : '/all?';
      const response = await fetch(`${endpoint}&f=${page * size}&s=${size}`);
      if (!response.ok) throw new Error('Failed to fetch');
      const data = await response.json();
      // Check if torrents is null and set results accordingly
      const torrents = data.torrents || [];
      setResults(torrents);
      // Check if the number of results is less than the page size
      setHasMoreResults(torrents.length === size);
    } catch (err) { 
      setError((err as Error).message);
    }
  };

  useEffect(() => {
    const ws = new WebSocket(`${window.location.protocol === 'https:' ? 'wss' : 'ws'}://${window.location.host}/ws`);

    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      setTotalCount(data.totalCount);
    };

    const fetchTrackers = async () => {
      try {
        const response = await fetch('/trackers');
        if (!response.ok) throw new Error('Failed to fetch trackers');
        const data = await response.json();
        setTrackers(data.trackers);
      } catch (err) {
        setError((err as Error).message);
      }
    };

    fetchTrackers();

    return () => {
      ws.close();
    };
  }, []); // Fetch total count and trackers only once when the component mounts

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
      setSelectedTorrent(data.torrents[0]);
      setShowModal(true); // Show the modal when a torrent is selected
    } catch (err) {
      setError((err as Error).message);
    }
  };

  const confirmDelete = (hash: string) => {
    setTorrentToDelete(hash);
    setShowDeleteModal(true);
  };

  const handleDelete = async () => {
    if (torrentToDelete) {
      try {
        const response = await fetch(`/delete?h=${torrentToDelete}`, { method: 'DELETE' });
        if (!response.ok) throw new Error('Failed to delete');
        // Refresh the results after deletion
        fetchResults();
        setShowDeleteModal(false); // Close the modal
        setShowModal(false); // Close the details modal
        setTorrentToDelete(null); // Reset the torrent to delete
      } catch (err) {
        setError((err as Error).message);
      }
    }
  };



  const TorrentDetailsModal = React.lazy(() => import('./TorrentDetailsModal'));
  const DeleteConfirmationModal = React.lazy(() => import('./DeleteConfirmationModal'));

  return (
    <div className="container mt-5">
      <h1 className="text-center mb-4">Welcome to Torrent Search</h1>
      {totalCount !== null && (
        <p className="text-center">Total Torrents Indexed: {totalCount}</p>
      )}
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
          torrent && (
            <li key={torrent.infohashHex} className="list-group-item d-flex justify-content-between align-items-center">
              <span className="torrent-name">{torrent.name}</span>
              <span className="torrent-size">{formatBytes(torrent.length)}</span>
              <div>
                <button className="btn btn-outline-primary" onClick={() => handleTorrent(torrent.infohashHex)}>Details</button>
              </div>
            </li>
          )
        ))}
      </ul>
      <div className="d-flex justify-content-between">
        <button className="btn btn-outline-primary" onClick={() => setPage((prev) => Math.max(prev - 1, 0))} disabled={page === 0}>
          Previous
        </button>
        <span>Page {page + 1}</span>
        <button
          className="btn btn-outline-primary"
          onClick={() => setPage((prev) => prev + 1)}
          disabled={!hasMoreResults}
        >
          Next
        </button>
      </div>
      <Suspense fallback={<div>Loading...</div>}>
        {selectedTorrent && (
          <TorrentDetailsModal
            selectedTorrent={selectedTorrent}
            showModal={showModal}
            setShowModal={setShowModal}
            confirmDelete={confirmDelete}
            trackers={trackers} // Pass trackers to the modal
          />
        )}
        {showDeleteModal && (
          <DeleteConfirmationModal
            showDeleteModal={showDeleteModal}
            setShowDeleteModal={setShowDeleteModal}
            handleDelete={handleDelete}
          />
        )}
      </Suspense>
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

import { formatBytes } from './utils';
