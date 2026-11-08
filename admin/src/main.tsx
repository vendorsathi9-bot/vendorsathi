import React from 'react';
import ReactDOM from 'react-dom/client';
import { App } from './app/App';
import { AuthProvider } from './core/auth/AuthContext';
import { LanguageProvider } from './core/i18n/LanguageContext';
import './index.css';

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <LanguageProvider>
      <AuthProvider>
        <App />
      </AuthProvider>
    </LanguageProvider>
  </React.StrictMode>
);
