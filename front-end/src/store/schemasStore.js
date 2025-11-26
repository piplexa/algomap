import { create } from 'zustand';
import { schemasAPI } from '../api/client';

export const useSchemasStore = create((set, get) => ({
  schemas: [],
  currentSchema: null,
  loading: false,
  error: null,

  fetchSchemas: async () => {
    set({ loading: true, error: null });
    try {
      const response = await schemasAPI.getAll();
      set({ schemas: response.data, loading: false });
    } catch (error) {
      set({ 
        error: error.response?.data?.message || 'Ошибка загрузки схем',
        loading: false 
      });
    }
  },

  fetchSchemaById: async (id) => {
    set({ loading: true, error: null });
    try {
      const response = await schemasAPI.getById(id);
      set({ currentSchema: response.data, loading: false });
    } catch (error) {
      set({ 
        error: error.response?.data?.message || 'Ошибка загрузки схемы',
        loading: false 
      });
    }
  },

  createSchema: async (data) => {
    set({ loading: true, error: null });
    try {
      const response = await schemasAPI.create(data);
      const newSchema = response.data;
      set((state) => ({
        schemas: [...state.schemas, newSchema],
        currentSchema: newSchema,
        loading: false,
      }));
      return { success: true, schema: newSchema };
    } catch (error) {
      set({ 
        error: error.response?.data?.message || 'Ошибка создания схемы',
        loading: false 
      });
      return { success: false };
    }
  },

  updateSchema: async (id, data) => {
    set({ loading: true, error: null });
    try {
      const response = await schemasAPI.update(id, data);
      const updatedSchema = response.data;
      set((state) => ({
        schemas: state.schemas.map((s) => (s.id === id ? updatedSchema : s)),
        currentSchema: updatedSchema,
        loading: false,
      }));
      return { success: true };
    } catch (error) {
      set({ 
        error: error.response?.data?.message || 'Ошибка обновления схемы',
        loading: false 
      });
      return { success: false };
    }
  },

  deleteSchema: async (id) => {
    set({ loading: true, error: null });
    try {
      await schemasAPI.delete(id);
      set((state) => ({
        schemas: state.schemas.filter((s) => s.id !== id),
        currentSchema: state.currentSchema?.id === id ? null : state.currentSchema,
        loading: false,
      }));
      return { success: true };
    } catch (error) {
      set({ 
        error: error.response?.data?.message || 'Ошибка удаления схемы',
        loading: false 
      });
      return { success: false };
    }
  },

  setCurrentSchema: (schema) => set({ currentSchema: schema }),
  clearError: () => set({ error: null }),
}));
