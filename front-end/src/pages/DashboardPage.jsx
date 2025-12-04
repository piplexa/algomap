import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useSchemasStore } from '../store/schemasStore';
import { useAuthStore } from '../store/authStore';
import '../styles/Dashboard.css';

// Статусы схемы из справочника dict_schema_status
const SCHEMA_STATUSES = {
  1: { name: 'draft', label: 'Черновик', description: 'схема в разработке' },
  2: { name: 'active', label: 'Активна', description: 'схема работает' },
  3: { name: 'archived', label: 'Архив', description: 'схема устарела' },
};

const getStatusLabel = (statusId) => {
  return SCHEMA_STATUSES[statusId]?.label || `Статус ${statusId}`;
};

export default function DashboardPage() {
  const navigate = useNavigate();
  const { logout } = useAuthStore();
  const { schemas, loading, error, fetchSchemas, deleteSchema } = useSchemasStore();
  const [showCreateModal, setShowCreateModal] = useState(false);

  useEffect(() => {
    fetchSchemas();
  }, [fetchSchemas]);

  const handleCreateSchema = () => {
    setShowCreateModal(true);
  };

  const handleDeleteSchema = async (id, name) => {
    if (window.confirm(`Удалить схему "${name}"?`)) {
      const result = await deleteSchema(id);
      if (result.success) {
        alert('Схема удалена');
      }
    }
  };

  return (
    <div className="dashboard-container">
      <header className="dashboard-header">
        <div className="header-left">
          <h1>🎨 AlgoMap</h1>
          <span className="header-subtitle">Ваши схемы</span>
        </div>
        <div className="header-right">
          <button onClick={handleCreateSchema} className="btn-primary">
            + Создать схему
          </button>
          <button onClick={logout} className="btn-secondary">
            Выход
          </button>
        </div>
      </header>

      <main className="dashboard-main">
        {loading && <div className="loading">Загрузка...</div>}
        {error && <div className="error-message">{error}</div>}

        {!loading && schemas.length === 0 && (
          <div className="empty-state">
            <p>У вас пока нет схем</p>
            <button onClick={handleCreateSchema} className="btn-primary">
              Создать первую схему
            </button>
          </div>
        )}

        <div className="schemas-grid">
          {schemas.map((schema) => (
            <div key={schema.id} className="schema-card">
              <div className="schema-card-header">
                <h3>{schema.name}</h3>
                <span className={`status-badge status-${schema.status}`}>
                  {getStatusLabel(schema.status)}
                </span>
              </div>

              {schema.description && (
                <p className="schema-description">{schema.description}</p>
              )}

              <div className="schema-meta">
                <span>📅 {new Date(schema.created_at).toLocaleDateString()}</span>
                <span>
                  🔗 {schema.definition?.nodes?.length || 0} нод
                </span>
              </div>

              <div className="schema-actions">
                <button
                  onClick={() => navigate(`/editor/${schema.id}`)}
                  className="btn-primary"
                >
                  Редактировать
                </button>
                <button
                  onClick={() => handleDeleteSchema(schema.id, schema.name)}
                  className="btn-danger"
                >
                  Удалить
                </button>
              </div>
            </div>
          ))}
        </div>
      </main>

      {showCreateModal && (
        <CreateSchemaModal
          onClose={() => setShowCreateModal(false)}
          onSuccess={(schema) => {
            setShowCreateModal(false);
            navigate(`/editor/${schema.id}`);
          }}
        />
      )}
    </div>
  );
}

function CreateSchemaModal({ onClose, onSuccess }) {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [loading, setLoading] = useState(false);
  const { createSchema } = useSchemasStore();

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);

    const result = await createSchema({
      name,
      description,
      definition: {
        nodes: [],
        edges: [],
      },
    });

    setLoading(false);

    if (result.success) {
      onSuccess(result.schema);
    } else {
      alert('Ошибка создания схемы');
    }
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()}>
        <h2>Создать новую схему</h2>
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>Название*</label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Моя первая схема"
              required
              autoFocus
            />
          </div>

          <div className="form-group">
            <label>Описание</label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Что делает эта схема?"
              rows={3}
            />
          </div>

          <div className="modal-actions">
            <button type="button" onClick={onClose} className="btn-secondary">
              Отмена
            </button>
            <button type="submit" className="btn-primary" disabled={loading}>
              {loading ? 'Создание...' : 'Создать'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
