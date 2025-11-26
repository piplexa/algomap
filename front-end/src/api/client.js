import axios from 'axios';

const api = axios.create({
  baseURL: '/api',
  headers: {
    'Content-Type': 'application/json',
  },
});

// Интерцептор для добавления токена в каждый запрос
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('session_key');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Интерцептор для обработки ошибок авторизации
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('session_key');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

// === Auth API ===
export const authAPI = {
  register: (email, password, name) =>
    api.post('/users/register', { email, password, name }),

  login: (email, password) =>
    api.post('/auth/login', { email, password }),

  logout: () =>
    api.post('/auth/logout'),
};

// === Schemas API ===
export const schemasAPI = {
  // TODO (backend): GET /schemas должен возвращать только схемы текущего пользователя
  // Backend должен использовать created_by из сессии для фильтрации
  getAll: () =>
    api.get('/schemas'),

  getById: (id) =>
    api.get(`/schemas/${id}`),

  create: (data) =>
    api.post('/schemas', data),

  update: (id, data) =>
    api.put(`/schemas/${id}`, data),

  delete: (id) =>
    api.delete(`/schemas/${id}`),
};

export default api;
