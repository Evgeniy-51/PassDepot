import React from 'react'
import {createRoot} from 'react-dom/client'
import './theme.css'
import './buttons.css'
import './password-input.css'
import './style.css'
import App from './App'
import { applyStoredTheme } from './theme'
import { applyStoredLocale } from './locale'

applyStoredTheme()
applyStoredLocale()

const container = document.getElementById('root')

const root = createRoot(container!)

root.render(
    <React.StrictMode>
        <App/>
    </React.StrictMode>
)
