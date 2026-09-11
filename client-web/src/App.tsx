// import { useState } from 'react';
import { useAuthStore } from './store/authStore';
import Login from './components/Login';
import ChatApp from './components/ChatApp';

function App() {
  const { isAuthenticated } = useAuthStore();

  return (
    <div className='app'>
      {isAuthenticated ? <ChatApp /> : <Login />}
    </div>
  );
}

export default App;