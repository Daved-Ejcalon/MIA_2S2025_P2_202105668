import React, { useState, useEffect } from 'react';
import './FileSystemVisualizer.css';

const FileSystemVisualizer = ({ isLoggedIn, sessionInfo }) => {
  const [disks, setDisks] = useState([]);
  const [selectedDisk, setSelectedDisk] = useState(null);
  const [selectedPartition, setSelectedPartition] = useState(null);
  const [currentPath, setCurrentPath] = useState('/');
  const [fileSystemContent, setFileSystemContent] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  // Emoji para identificar los discos
  const diskEmoji = '💿';

  useEffect(() => {
    if (isLoggedIn) {
      loadDisks();
    }
  }, [isLoggedIn]);

  const loadDisks = async () => {
    try {
      setLoading(true);
      const response = await fetch('http://localhost:8080/disks', {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      const result = await response.json();

      if (result.error) {
        setError(result.error);
      } else {
        setDisks(result.disks || []);
      }
    } catch (err) {
      setError(`Error de conexión: ${err.message}`);
    } finally {
      setLoading(false);
    }
  };

  const handleDiskClick = (disk) => {
    setSelectedDisk(disk);
    setSelectedPartition(null);
  };

  const handlePartitionClick = (partition) => {
    setSelectedPartition(partition);
    setCurrentPath('/');
    loadFileSystemContent(partition.id, '/');
  };

  const handleBackToPartitions = () => {
    setSelectedPartition(null);
    setCurrentPath('/');
    setFileSystemContent([]);
  };

  const loadFileSystemContent = async (partitionId, path) => {
    try {
      const response = await fetch(`http://localhost:8080/filesystem?partition_id=${partitionId}&path=${encodeURIComponent(path)}`);

      if (!response.ok) {
        const errorData = await response.json();
        console.error('Error del servidor:', errorData.error);
        setFileSystemContent([]);
        return;
      }

      const data = await response.json();
      // Asegurar que data sea un array, si es null o undefined usar array vacío
      setFileSystemContent(Array.isArray(data) ? data : []);
    } catch (err) {
      console.error('Error al cargar contenido:', err);
      setFileSystemContent([]);
    }
  };

  const handleFileItemClick = (item) => {
    if (item.type === 'folder') {
      // Verificar si ya estamos en este directorio para evitar duplicación
      const pathParts = currentPath.split('/').filter(p => p);
      const lastPart = pathParts[pathParts.length - 1];

      // Si el último componente del path es el mismo que el nombre del item, no hacer nada
      if (lastPart === item.name) {
        return;
      }

      const newPath = currentPath === '/' ? `/${item.name}` : `${currentPath}/${item.name}`;
      setCurrentPath(newPath);
      loadFileSystemContent(selectedPartition.id, newPath);
    } else {
      alert(`Archivo: ${item.name}\nTamaño: ${formatBytes(item.size)}\nPermisos: ${item.permissions}`);
    }
  };

  const goToParentFolder = () => {
    if (currentPath === '/') return;
    const parts = currentPath.split('/').filter(p => p);
    parts.pop();
    const newPath = parts.length === 0 ? '/' : '/' + parts.join('/');
    setCurrentPath(newPath);
    loadFileSystemContent(selectedPartition.id, newPath);
  };

  const formatBytes = (bytes) => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
  };

  const formatSize = (bytes) => {
    const mb = bytes / (1024 * 1024);
    if (mb >= 1024) {
      return `${(mb / 1024).toFixed(2)} GB`;
    }
    return `${mb.toFixed(2)} MB`;
  };

  if (!isLoggedIn) {
    return (
      <div className="visualizer-container">
        <div className="not-logged-in">
          <h2>Acceso Restringido</h2>
          <p>Debes iniciar sesión para acceder al visualizador del sistema de archivos.</p>
        </div>
      </div>
    );
  }

  if (loading) {
    return (
      <div className="visualizer-container">
        <div className="loading">
          <h2>Cargando discos...</h2>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="visualizer-container">
        <div className="error-message">
          <h2>Error</h2>
          <p>{error}</p>
          <button onClick={loadDisks} className="retry-btn">Reintentar</button>
        </div>
      </div>
    );
  }

  return (
    <div className="visualizer-container">
      <div className="visualizer-header">
        <h1>Visualizador del Sistema de Archivos</h1>
        {sessionInfo && (
          <div className="session-info">
            <span>Usuario: {sessionInfo.username}</span>
            <span>Partición: {sessionInfo.mountID}</span>
          </div>
        )}
      </div>

      {!selectedDisk ? (
        <div className="disk-selection">
          <h2>Seleccione el disco que desea visualizar:</h2>

          {disks.length === 0 ? (
            <div className="no-disks">
              <p>No hay discos disponibles. Cree discos usando el comando <code>mkdisk</code>.</p>
            </div>
          ) : (
            <div className="disk-grid">
              {disks.map((disk, index) => (
                <div
                  key={index}
                  className="disk-card"
                  onClick={() => handleDiskClick(disk)}
                >
                  <div className="disk-icon">
                    {diskEmoji}
                  </div>
                  <div className="disk-name">{disk.name}</div>
                  <div className="disk-info">
                    <div className="info-item">
                      <span className="label">Capacidad:</span>
                      <span className="value">{formatSize(disk.size)}</span>
                    </div>
                    <div className="info-item">
                      <span className="label">Fit:</span>
                      <span className="value">{disk.fit}</span>
                    </div>
                    <div className="info-item">
                      <span className="label">Particiones:</span>
                      <span className="value">{disk.mountedPartitions} montada(s)</span>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      ) : (
        <div className="disk-explorer">
          <div className="explorer-header">
            <div className="header-left">
              <button
                className="back-btn"
                onClick={() => selectedPartition ? handleBackToPartitions() : setSelectedDisk(null)}
              >
                {selectedPartition ? 'Volver a Particiones' : 'Volver a selección de discos'}
              </button>
              <h2>{selectedDisk.name}</h2>
            </div>
            {selectedPartition && (
              <div className="partition-info-text">
                <span>Partición: {selectedPartition.name}</span>
                <span>ID: {selectedPartition.id}</span>
              </div>
            )}
          </div>

          {!selectedPartition ? (
            <div className="partitions-section">
              <h3>Particiones del Disco</h3>
              {selectedDisk.partitions && selectedDisk.partitions.length > 0 ? (
                <div className="partitions-list">
                  {selectedDisk.partitions.map((partition, idx) => (
                    <div
                      key={idx}
                      className={`partition-card ${partition.isMounted ? 'mounted' : 'unmounted'}`}
                      onClick={() => partition.isMounted && handlePartitionClick(partition)}
                      style={{ cursor: partition.isMounted ? 'pointer' : 'default' }}
                    >
                      <div className="partition-header">
                        <span className="partition-name">{partition.name}</span>
                        <span className={`partition-status ${partition.isMounted ? 'status-mounted' : 'status-unmounted'}`}>
                          {partition.isMounted ? 'Montada' : 'No Montada'}
                        </span>
                      </div>
                      {partition.isMounted && partition.id && (
                        <div className="partition-id-badge">
                          ID: {partition.id}
                        </div>
                      )}
                      <div className="partition-details">
                        <div className="detail-row">
                          <span className="detail-label">Tipo:</span>
                          <span className="detail-value">{partition.type}</span>
                        </div>
                        <div className="detail-row">
                          <span className="detail-label">Tamaño:</span>
                          <span className="detail-value">{formatSize(partition.size)}</span>
                        </div>
                        <div className="detail-row">
                          <span className="detail-label">Ajuste:</span>
                          <span className="detail-value">{partition.fit}</span>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="no-partitions">No hay particiones en este disco.</p>
              )}
            </div>
          ) : (
            <div className="file-explorer-section">
              <div className="breadcrumb-bar">
                {currentPath === '/' ? (
                  <button className="breadcrumb-btn active" disabled>
                    / (Raíz)
                  </button>
                ) : (
                  <>
                    <button className="breadcrumb-btn" onClick={() => { setCurrentPath('/'); loadFileSystemContent(selectedPartition.id, '/'); }}>
                      /
                    </button>
                    <span className="breadcrumb-separator">/</span>
                    <button className="breadcrumb-btn" onClick={goToParentFolder}>Atrás</button>
                    <span className="current-path-text">{currentPath}</span>
                  </>
                )}
              </div>

              <div className="file-list-container">
                {fileSystemContent.length === 0 ? (
                  <div className="empty-folder">
                    <p>Esta carpeta está vacía</p>
                  </div>
                ) : (
                  <div className="file-grid">
                    {fileSystemContent.map((item, index) => (
                      <div
                        key={index}
                        className={`file-item ${item.type === 'folder' ? 'folder-item' : 'file-item-type'}`}
                        onClick={() => handleFileItemClick(item)}
                      >
                        <div className="file-icon">
                          {item.type === 'folder' ? '📁' : '📄'}
                        </div>
                        <div className="file-info">
                          <div className="file-name">{item.name}</div>
                          {item.type === 'file' && (
                            <div className="file-size">{formatBytes(item.size)}</div>
                          )}
                          <div className="file-permissions">#{item.permissions}</div>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default FileSystemVisualizer;
