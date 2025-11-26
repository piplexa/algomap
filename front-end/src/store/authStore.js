import { create } from 'zustand';
import { authAPI } from '../api/client';

export const useAuthStore = create((set) => ({
  isAuthenticated: !!localStorage.getItem('session_key'),
  user: null,

  login: async (email, password) => {
    try {
      const response = await authAPI.login(email, password);
      const { session_key, user } = response.data;
      
      localStorage.setItem('session_key', session_key);
      set({ isAuthenticated: true, user });
      
      return { success: true };
    } catch (error) {
      return { 
        success: false, 
        error: error.response?.data?.message || 'Ошибка входа' 
      };
    }
  },

  register: async (email, password, name) => {
    try {
      await authAPI.register(email, password, name);
      return { success: true };
    } catch (error) {
      return { 
        success: false, 
        error: error.response?.data?.message || 'Ошибка регистрации' 
      };
    }
  },

  logout: async () => {
    try {
      await authAPI.logout();
    } catch (error) {
      console.error('Logout error:', error);
    } finally {
      localStorage.removeItem('session_key');
      set({ isAuthenticated: false, user: null });
    }
  },
}));
