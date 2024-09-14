import React, { useState, useEffect, Suspense } from 'react';
import 'bootstrap/dist/css/bootstrap.min.css';
import { createRoot } from 'react-dom/client';

const App: React.FC = () => {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState([]);
  const [error, setError] = useState('');
  const [page, setPage] = useState(0); // Current page
  const [size, setSize] = useState(10); // Results per page
  const [totalCount, setTotalCount] = useState<number | null>(null);
  const [isSearching, setIsSearching] = useState(false); // Track if search is active
  const [hasMoreResults, setHasMoreResults] = useState(true);
  const [selectedTorrent, setSelectedTorrent] = useState<any | null>(null);
  const [showModal, setShowModal] = useState(false); // State to control modal visibility
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [torrentToDelete, setTorrentToDelete] = useState<string | null>(null);

  const fetchResults = async () => {
    try {
      const endpoint = isSearching ? `/query?q=${query}` : '/all?';
      const response = await fetch(`${endpoint}&f=${page * size}&s=${size}`);
      if (!response.ok) throw new Error('Failed to fetch');
      const data = await response.json();
      setResults(data.torrents);
      // Check if the number of results is less than the page size
      setHasMoreResults(data.torrents.length === size);
    } catch (err) { 
      setError((err as Error).message);
    }
  };

  useEffect(() => {
    const fetchTotalCount = async () => {
      try {
        const response = await fetch('/count');
        if (!response.ok) throw new Error('Failed to fetch total count');
        const data = await response.json();
        setTotalCount(data.totalCount);
      } catch (err) {
        setError((err as Error).message);
      }
    };

    fetchTotalCount();
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



  const FileTree = React.lazy(() => import('./FileTree'));

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
          <li key={torrent.infohashHex} className="list-group-item d-flex justify-content-between align-items-center">
            {torrent.name} - {formatBytes(torrent.length)}
            <div>
              <button className="btn btn-link" onClick={() => handleTorrent(torrent.infohashHex)}>Details</button>
            </div>
          </li>
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
      {/* Modal for Torrent Details */}
      {selectedTorrent && (
        <div className={`modal ${showModal ? 'd-block' : 'd-none'}`} tabIndex={-1} role="dialog">
          <div className="modal-dialog" role="document">
            <div className="modal-content">
              <div className="modal-header">
                <h5 className="modal-title">Torrent Details</h5>
                <button type="button" className="close" onClick={() => setShowModal(false)} aria-label="Close">
                  <span aria-hidden="true">&times;</span>
                </button>
              </div>
              <div className="modal-body">
                <p>Name: {selectedTorrent.name}</p>
                <p>Size: {formatBytes(selectedTorrent.length)}</p>
                <p>
                  Link: 
                  <a href={`magnet:?xt=urn:btih:${selectedTorrent.infohashHex}`} target="_blank" rel="noopener noreferrer">
                    {"🧲"}
                  </a>
                </p>
                <h3>Files:</h3>
                <Suspense fallback={<div>Loading files...</div>}>
                  <FileTree files={selectedTorrent.files} />
                </Suspense>
              </div>
              <div className="modal-footer">
                <button type="button" className="btn btn-danger" onClick={() => confirmDelete(selectedTorrent.infohashHex)}>Delete</button>
                <button type="button" className="btn btn-secondary" onClick={() => setShowModal(false)}>Close</button>
              </div>
            </div>
          </div>
        </div>
      )}
      {/* Modal for Delete Confirmation */}
      {showDeleteModal && (
        <div className="modal d-block" tabIndex={-1} role="dialog">
          <div className="modal-dialog" role="document">
            <div className="modal-content">
              <div className="modal-header">
                <h5 className="modal-title">Confirm Deletion</h5>
                <button type="button" className="close" onClick={() => setShowDeleteModal(false)} aria-label="Close">
                  <span aria-hidden="true">&times;</span>
                </button>
              </div>
              <div className="modal-body">
                <p>Are you sure you want to delete this torrent?</p>
              </div>
              <div className="modal-footer">
                <button type="button" className="btn btn-secondary" onClick={() => setShowDeleteModal(false)}>Cancel</button>
                <button type="button" className="btn btn-danger" onClick={handleDelete}>Delete</button>
              </div>
            </div>
          </div>
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

export const formatBytes = (bytes: number, decimals = 2) => {
  if (bytes === 0) return '0 Bytes';
  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
};