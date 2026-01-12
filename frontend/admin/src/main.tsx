import { render } from 'preact';
import { initializeTheme } from '@/signals/theme';
import { App } from './app';
import './index.css';

// Initialize theme before rendering
initializeTheme();

render(<App />, document.getElementById('app')!);
