import React, { useState, useEffect } from 'react';
import './FileSystemVisualizer.css';

const FileSystemVisualizer = ({ isLoggedIn, sessionInfo }) => {
  const [disks, setDisks] = useState([]);
  const [selectedDisk, setSelectedDisk] = useState(null);
  const [selectedPartition, setSelectedPartition] = useState(null);
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
  };

  const handleBackToPartitions = () => {
    setSelectedPartition(null);
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
            <button
              className="back-btn"
              onClick={() => setSelectedDisk(null)}
            >
              Volver a selección de discos
            </button>
            <h2>{selectedDisk.name}</h2>
          </div>

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
        </div>
      )}
    </div>
  );
};

export default FileSystemVisualizer;
