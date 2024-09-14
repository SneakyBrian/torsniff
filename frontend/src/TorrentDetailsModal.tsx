import React, { Suspense } from 'react';
import { formatBytes } from './utils';

const FileTree = React.lazy(() => import('./FileTree'));

interface TorrentDetailsModalProps {
  selectedTorrent: any;
  showModal: boolean;
  setShowModal: (show: boolean) => void;
  confirmDelete: (hash: string) => void;
}

const TorrentDetailsModal: React.FC<TorrentDetailsModalProps> = ({
  selectedTorrent,
  showModal,
  setShowModal,
  confirmDelete,
}) => {
  return (
    <div className={`modal ${showModal ? 'd-block' : 'd-none'}`} tabIndex={-1} role="dialog">
      <div className="modal-dialog" role="document">
        <div className="modal-content">
          <div className="modal-header">
            <h5 className="modal-title">Torrent Details</h5>
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
            <div className="files-section">
              <Suspense fallback={<div>Loading files...</div>}>
                <FileTree files={selectedTorrent.files} />
              </Suspense>
            </div>
          </div>
          <div className="modal-footer">
            <button type="button" className="btn btn-danger" onClick={() => confirmDelete(selectedTorrent.infohashHex)}>Delete</button>
            <button type="button" className="btn btn-secondary" onClick={() => setShowModal(false)}>Close</button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default TorrentDetailsModal;
