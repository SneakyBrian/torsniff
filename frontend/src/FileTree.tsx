import React, { useState } from 'react';
import { formatBytes } from './index';

const FileTree: React.FC<{ files: any[] }> = ({ files }) => {
  const [expandedFolders, setExpandedFolders] = useState<Set<string>>(new Set());

  const toggleFolder = (path: string) => {
    setExpandedFolders((prev) => {
      const newSet = new Set(prev);
      if (newSet.has(path)) {
        newSet.delete(path);
      } else {
        newSet.add(path);
      }
      return newSet;
    });
  };

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
      const currentPath = [...path, key].join('/');
      const isExpanded = expandedFolders.has(currentPath);

      if (typeof value === 'number') {
        return (
          <li key={currentPath}>
            {key} - {formatBytes(value)}
          </li>
        );
      }

      return (
        <li key={currentPath}>
          <span onClick={() => toggleFolder(currentPath)} style={{ cursor: 'pointer' }}>
            {isExpanded ? '▼' : '▶'} <strong>{key}</strong>
          </span>
          {isExpanded && <ul>{renderTree(value, [...path, key])}</ul>}
        </li>
      );
    });
  };

  return <ul>{renderTree(fileTree)}</ul>;
};

export default FileTree;
